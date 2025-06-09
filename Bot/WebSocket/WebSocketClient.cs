using NativeWebSocket;
using System.Text;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using Websocket.Models;

namespace Websocket.Services;

public class WebSocketClient : IWebSocketClient
{
    private readonly ILogger<WebSocketClient> _logger;
    private MessageHandlerRegistry MessageHandlerRegistry;
    private WebSocket _webSocket;
    private readonly CancellationTokenSource _cancellationTokenSource = new();
    private string _playerId;
    private string _wsUrl;

    public event Func<string, Task> OnError;

    public string PlayerId => _playerId;

    public WebSocketClient(ILogger<WebSocketClient> logger)
    {
        _logger = logger;
    }

    #region Core
    public async Task StartAsync(string playerId, string token, string wsUrl)
    {
        try
        {
            _playerId = playerId;
            _wsUrl = wsUrl;

            var headers = new Dictionary<string, string> { { "Authorization", $"Bearer {token}" } };
            _webSocket = new WebSocket($"ws://{_wsUrl}/ws?token={token}", headers);

            _webSocket.OnOpen += () =>
            {
                _logger.LogInformation("Connected to WebSocket server via WebSocket");
            };

            _webSocket.OnMessage += async bytes =>
            {
                var messageJson = Encoding.UTF8.GetString(bytes);
                await HandleMessageAsync(messageJson);
            };

            _webSocket.OnError += async errorMsg =>
            {
                _logger.LogError($"WebSocket error: {errorMsg}");
                if (OnError != null)
                    await OnError.Invoke(errorMsg);
            };

            _webSocket.OnClose += async closeCode =>
            {
                _logger.LogInformation($"WebSocket connection closed with code: {closeCode}");
            };

            await _webSocket.Connect();
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to start WebSocket client with NativeWebSocket");
            if (OnError != null)
                await OnError.Invoke(ex.Message);
            throw;
        }
    }

    private async Task SendMessageAsync<T>(Message<T> message)
    {
        try
        {
            if (_webSocket == null || _webSocket.State != WebSocketState.Open)
            {
                _logger.LogWarning("WebSocket is not open. Cannot send message.");
                if (OnError != null)
                    await OnError.Invoke("WebSocket is not open. Cannot send message. Type:" + message.Type);

                return;
            }

            var json = JsonConvert.SerializeObject(message);
            await _webSocket.SendText(json);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error sending message via NativeWebSocket");
            if (OnError != null)
                await OnError.Invoke(ex.Message);
        }
    }

    private async Task HandleMessageAsync(string messageJson)
    {
        try
        {
            var message = JsonConvert.DeserializeObject<Response<object>>(messageJson);
            _logger.LogInformation("Received message: {Message}", messageJson);
            if (message == null)
            {
                _logger.LogError("Received null message");
                return;
            }

            _logger.LogInformation("Message type: {MessageType}", message.Type);
            await MessageHandlerRegistry.HandleMessageAsync(message.Type, message);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error handling message: {Message}", messageJson);
            if (OnError != null)
                await OnError.Invoke(ex.Message);
        }
    }
    #endregion

    #region Actions
    public async Task JoinRoomAsync(string roomId)
    {
        var message = new Message<MessageRoomJoin>
        {
            Type = MessageType.RoomJoin,
            Data = new MessageRoomJoin
            {
                RoomId = roomId,
                PlayerId = _playerId
            }
        };

        await SendMessageAsync(message);
    }

    public async Task LeaveRoomAsync(string roomId)
    {
        var message = new Message<MessageRoomLeave>
        {
            Type = MessageType.RoomLeave,
            Data = new MessageRoomLeave
            {
                RoomId = roomId,
                PlayerId = _playerId
            }
        };

        await SendMessageAsync(message);
    }

    public async Task JoinGameAsync(string roomId, int position)
    {
        var message = new Message<MessageGameJoin>
        {
            Type = MessageType.GameJoin,
            Data = new MessageGameJoin
            {
                RoomId = roomId,
                PlayerId = _playerId,
                Position = position
            }
        };

        await SendMessageAsync(message);
    }

    public async Task LeaveGameAsync(string roomId)
    {
        var message = new Message<MessageGameLeave>
        {
            Type = MessageType.GameLeave,
            Data = new MessageGameLeave
            {
                RoomId = roomId,
                PlayerId = _playerId
            }
        };

        await SendMessageAsync(message);
    }

    public async Task SendGameActionAsync<T>(string roomId, T action)
    {
        if (action == null) throw new ArgumentNullException(nameof(action));

        var message = new Message<MessageGameAction>
        {
            Type = MessageType.GameAction,
            Data = new MessageGameAction
            {
                RoomId = roomId,
                PlayerId = _playerId,
                Data = JToken.FromObject(action)
            }
        };

        await SendMessageAsync(message);
    }
    #endregion

    #region Http
    public async Task<IEnumerable<RoomSummary>> GetRoomsByGameTypeAsync(GameType gameType)
    {
        var client = new HttpClient();
        var response = await client.GetAsync($"http://{_wsUrl}/rooms?game_type={(int)gameType}");
        var content = await response.Content.ReadAsStringAsync();
        var rooms = JsonConvert.DeserializeObject<IEnumerable<RoomSummary>>(content);

        return rooms ?? [];
    }
    #endregion

    public void SetMessageHandlerRegistry(MessageHandlerRegistry messageHandlerRegistry)
    {
        MessageHandlerRegistry = messageHandlerRegistry;
    }

    public async Task CloseConnectionAsync()
    {
        _cancellationTokenSource.Cancel();

        if (_webSocket != null && (_webSocket.State == NativeWebSocket.WebSocketState.Open || _webSocket.State == NativeWebSocket.WebSocketState.Connecting))
        {
            _logger.LogInformation("Closing WebSocket connection (NativeWebSocket)");
            await _webSocket.Close();
        }
        _cancellationTokenSource.Dispose();
    }

    public void Dispose()
    {
        CloseConnectionAsync().ConfigureAwait(false).GetAwaiter().GetResult();
    }
}

public class MessageHandlerRegistry
{
    private readonly Dictionary<string, List<Func<object, DateTime, Task>>> _handlers = [];

    public MessageHandlerRegistry()
    {
    }

    public void On<T>(string eventName, Func<T, DateTime, Task> handler)
    {
        if (!_handlers.TryGetValue(eventName, out List<Func<object, DateTime, Task>>? value))
        {
            value ??= [];
            _handlers[eventName] = value;
        }

        value.Add(async (data, timestamp) =>
        {
            try
            {
                if (data is JToken element)
                {
                    var typedData = element.ToObject<T>();
                    if (typedData != null) await handler(typedData, timestamp);
                }
                else
                {
                    var typedData = JsonConvert.DeserializeObject<T>(JsonConvert.SerializeObject(data));
                    if (typedData != null) await handler(typedData, timestamp);
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Error] handling message of type {eventName}: {ex.Message}");
                throw;
            }
        });
    }

    public async Task HandleMessageAsync(string messageType, Response<object> data)
    {
        if (_handlers.TryGetValue(messageType, out var handlers))
        {
            foreach (var handler in handlers)
            {
                await handler(data.Data, data.Timestamp);
            }
        }
        else
        {
            Console.WriteLine($"[Warning] No handler registered for message type: {messageType}");
        }
    }
}