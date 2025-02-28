using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Bot.Models;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace Bot.Services;

public class WebSocketClientService : IWebSocketClientService
{
    private readonly IConfiguration _configuration;
    private readonly ILogger<WebSocketClientService> _logger;
    private readonly IPokerGameService _pokerGameService;
    private readonly IAuthService _authService;
    private readonly ClientWebSocket _webSocket;
    private readonly CancellationTokenSource _cancellationTokenSource;
    private Game _currentGame;
    private string _playerId;
    private string _token;

    public event Func<Game, Task> OnGameStateChanged;
    public event Func<string, Task> OnError;

    public WebSocketClientService(
        IConfiguration configuration,
        ILogger<WebSocketClientService> logger,
        IPokerGameService pokerGameService,
        IAuthService authService)
    {
        _configuration = configuration;
        _logger = logger;
        _pokerGameService = pokerGameService;
        _authService = authService;
        _webSocket = new ClientWebSocket();
        _cancellationTokenSource = new CancellationTokenSource();
    }

    public async Task StartAsync()
    {
        try
        {
            // Authenticate first
            var email = _configuration["Email"] ?? throw new Exception("Email not configured");
            var password = _configuration["Password"] ?? throw new Exception("Password not configured");

            _logger.LogInformation("Authenticating with email {Email}", email);
            var loginResponse = await _authService.LoginAsync(email, password);
            _token = loginResponse.Token;
            _playerId = loginResponse.User.Player.Id;

            // Get player details
            var player = await _authService.GetPlayerAsync(_token);
            _logger.LogInformation("Authenticated as player {PlayerName} with {Chips} chips", player.Username, player.Chips);

            // Connect to WebSocket with auth token
            var wsUrl = _configuration["WebSocketUrl"] ?? "ws://localhost:8080/ws";
            _webSocket.Options.SetRequestHeader("Authorization", $"Bearer {_token}");
            await _webSocket.ConnectAsync(new Uri(wsUrl), _cancellationTokenSource.Token);
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

    public async Task JoinGameAsync(string roomId)
    {
        var message = new Message
        {
            Type = MessageType.JoinGame,
            Data = new MessageJoinGame
            {
                RoomId = roomId,
                PlayerId = _playerId
            }
        };

        await SendMessageAsync(message);
    }

    public async Task LeaveGameAsync(string roomId)
    {
        var message = new Message
        {
            Type = MessageType.LeaveGame,
            Data = new MessageLeaveGame
            {
                RoomId = roomId,
                PlayerId = _playerId
            }
        };

        await SendMessageAsync(message);
    }

    public async Task SendGameActionAsync(GameAction action)
    {
        if (_currentGame == null)
        {
            _logger.LogWarning("Cannot send game action: no active game");
            return;
        }

        var message = new Message
        {
            Type = MessageType.GameAction,
            Data = new MessageGameAction
            {
                RoomId = _currentGame.Id,
                PlayerId = _playerId,
                Data = action
            }
        };

        await SendMessageAsync(message);
    }

    public Task<Game> GetGameStateAsync(string gameId)
    {
        return Task.FromResult(_currentGame);
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

    private async Task HandleMessageAsync(string messageJson)
    {
        try
        {
            var message = JsonSerializer.Deserialize<Message>(messageJson);
            _logger.LogInformation("Received message: {Message}", messageJson);
            if (message == null)
            {
                _logger.LogError("Received null message");
                return;
            }

            _logger.LogInformation("Message type: {MessageType}", message.Type);
            switch (message.Type)
            {
                case MessageType.GameInfo:
                    var room = JsonSerializer.Deserialize<Room>(message.Data.ToString());
                    if (room?.Game != null)
                    {
                        _currentGame = room.Game;
                        if (OnGameStateChanged != null)
                            await OnGameStateChanged.Invoke(_currentGame);

                        // If it's our turn and the game is started, determine and send the next action
                        if (_currentGame.Status == "started" && _pokerGameService.IsPlayerTurn(_currentGame, _playerId))
                        {
                            var action = await _pokerGameService.DetermineNextAction(_currentGame, _playerId);
                            if (action != null)
                                await SendGameActionAsync(action);
                        }
                    }
                    break;

                case MessageType.Error:
                    var error = message.Data.ToString();
                    _logger.LogError("Received error: {Error}", error);
                    if (OnError != null)
                        await OnError.Invoke(error);
                    break;
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error handling message: {Message}", messageJson);
            if (OnError != null)
                await OnError.Invoke(ex.Message);
        }
    }

    private async Task SendMessageAsync(Message message)
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

    public void Dispose()
    {
        _cancellationTokenSource.Cancel();
        _webSocket.Dispose();
        _cancellationTokenSource.Dispose();
    }
}