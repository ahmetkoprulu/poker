using System.Drawing;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Logging;
using Pastel;
using Websocket.Models;

namespace Websocket.Services;

public class WebSocketClient : IWebSocketClient
{
    private readonly ILogger<WebSocketClient> _logger;
    private MessageHandlerRegistry MessageHandlerRegistry;
    private readonly ClientWebSocket _webSocket = new();
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
            _webSocket.Options.SetRequestHeader("Authorization", $"Bearer {token}");
            await _webSocket.ConnectAsync(new Uri($"ws://{wsUrl}/ws"), _cancellationTokenSource.Token);
            _logger.LogInformation("Connected to WebSocket server");

            // Start receiving messages
            _ = ReceiveMessagesAsync();
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to start WebSocket client");
            throw;
        }
    }

    private async Task SendMessageAsync<T>(Message<T> message)
    {
        try
        {
            var json = JsonSerializer.Serialize(message);
            var buffer = Encoding.UTF8.GetBytes(json);
            await _webSocket.SendAsync(
                new ArraySegment<byte>(buffer),
                WebSocketMessageType.Text,
                true,
                _cancellationTokenSource.Token);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error sending message");
            if (OnError != null)
                await OnError.Invoke(ex.Message);
        }
    }

    private async Task ReceiveMessagesAsync()
    {
        var buffer = new byte[4096];

        try
        {
            while (_webSocket.State == WebSocketState.Open)
            {
                var result = await _webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), _cancellationTokenSource.Token);

                if (result.MessageType == WebSocketMessageType.Close)
                {
                    await _webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, string.Empty, _cancellationTokenSource.Token);
                    break;
                }

                var message = Encoding.UTF8.GetString(buffer, 0, result.Count);
                await HandleMessageAsync(message);
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error in WebSocket receive loop");
            if (OnError != null)
                await OnError.Invoke(ex.Message);
        }
    }

    private void SetMessageHandlerRegistry(MessageHandlerRegistry messageHandlerRegistry)
    {
        MessageHandlerRegistry = messageHandlerRegistry;
    }

    private async Task HandleMessageAsync(string messageJson)
    {
        try
        {
            var message = JsonSerializer.Deserialize<Message<object>>(messageJson);
            _logger.LogInformation("Received message: {Message}", messageJson);
            if (message == null)
            {
                _logger.LogError("Received null message");
                return;
            }

            _logger.LogInformation("Message type: {MessageType}", message.Type);
            await MessageHandlerRegistry.HandleMessageAsync(message.Type, message.Data);
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
        var message = new Message<MessageGameAction<T>>
        {
            Type = MessageType.GameAction,
            Data = new MessageGameAction<T>
            {
                RoomId = roomId,
                PlayerId = _playerId,
                Data = action
            }
        };

        await SendMessageAsync(message);
    }
    #endregion

    #region Http
    public async Task<IEnumerable<RoomSummary>> GetRoomsByGameTypeAsync(GameType gameType)
    {
        var client = new HttpClient();
        var response = await client.GetAsync($"http://{_wsUrl}/rooms?game_type={gameType}");
        var content = await response.Content.ReadAsStringAsync();
        var rooms = JsonSerializer.Deserialize<IEnumerable<RoomSummary>>(content);

        return rooms ?? [];
    }
    #endregion

    public void Dispose()
    {
        _cancellationTokenSource.Cancel();
        _webSocket.Dispose();
        _cancellationTokenSource.Dispose();
    }
}

public class MessageHandlerRegistry
{
    private readonly Dictionary<string, List<Func<object, Task>>> _handlers = [];

    public MessageHandlerRegistry()
    {
    }

    public void On<T>(string eventName, Func<T, Task> handler)
    {
        if (!_handlers.TryGetValue(eventName, out List<Func<object, Task>>? value))
        {
            value ??= [];
            _handlers[eventName] = value;
        }

        value.Add(async (data) =>
        {
            try
            {
                if (data is JsonElement element)
                {
                    var typedData = element.Deserialize<T>();
                    if (typedData != null) await handler(typedData);
                }
                else
                {
                    await handler((T)data);
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Error] handling message of type {eventName}: {ex.Message}".Pastel(Color.Red));
                throw;
            }
        });
    }

    public async Task HandleMessageAsync(string messageType, object data)
    {
        if (_handlers.TryGetValue(messageType, out var handlers))
        {
            foreach (var handler in handlers)
            {
                await handler(data);
            }
        }
        else
        {
            Console.WriteLine($"[Warning] No handler registered for message type: {messageType}".Pastel(Color.Yellow));
        }
    }
}