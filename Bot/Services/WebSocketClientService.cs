using System.Net.WebSockets;
using System.Text.Json;
using Bot.Models;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;
using Websocket.Client;

namespace Bot.Services;

public class WebSocketClientService : IWebSocketClientService
{
    private readonly WebsocketClient _client;
    private readonly UserPlayer _player;
    private readonly string _playerId;
    private Game _currentGameState;
    private string _currentRoomId;
    private string _currentGameId;
    private readonly JsonSerializerSettings _jsonOptions;

    public WebSocketClientService(string serverUrl, string token, UserPlayer userPlayer)
    {
        _player = userPlayer;
        _playerId = userPlayer.Id;
        _jsonOptions = new JsonSerializerSettings
        {
            ContractResolver = new DefaultContractResolver
            {
                NamingStrategy = new CamelCaseNamingStrategy()
            }
        };

        var url = new Uri(serverUrl);
        var factory = new Func<ClientWebSocket>(() =>
        {
            var client = new ClientWebSocket();
            client.Options.SetRequestHeader("Authorization", $"Bearer {token}");
            return client;
        });

        _client = new WebsocketClient(url, factory);
        _client.MessageReceived.Subscribe(msg =>
        {
            try
            {
                HandleMessage(msg.Text);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error handling message: {ex.Message}");
            }
        });
    }

    public async Task StartAsync()
    {
        await _client.Start();
    }

    public async Task JoinGameAsync(string roomId)
    {
        _currentRoomId = roomId;
        var message = new Message
        {
            Type = MessageType.JoinGame,
            RoomId = roomId,
            PlayerId = _playerId,
            Data = new { }
        };

        await SendMessageAsync(message);
    }

    public async Task SendGameActionAsync(string action)
    {
        if (string.IsNullOrEmpty(_currentGameId))
        {
            Console.WriteLine("No active game.");
            return;
        }

        var message = new Message
        {
            Type = MessageType.GameAction,
            GameId = _currentGameId,
            RoomId = _currentRoomId,
            PlayerId = _playerId,
            Data = new { action }
        };

        await SendMessageAsync(message);
    }

    private async Task SendMessageAsync(Message message)
    {
        var json = JsonConvert.SerializeObject(message, _jsonOptions);
        Console.WriteLine($"Sending message: {json}");
        await _client.SendInstant(json);
    }

    private void HandleMessage(string messageJson)
    {
        Console.WriteLine($"Received message: {messageJson}");
        var message = JsonConvert.DeserializeObject<Message>(messageJson, _jsonOptions) ?? throw new Exception("Message is null");

        switch (message.Type)
        {
            case MessageType.GameState:
                var gameStateJson = JsonConvert.SerializeObject(message.Data);
                var data = JsonConvert.DeserializeObject<Room>(gameStateJson, _jsonOptions) ?? throw new Exception("Game state is null");
                _currentGameState = data.Game;
                _currentGameId = message.GameId;

                // If we have 2 or more players and game is in waiting state, start the game
                if (_currentGameState.Players.Count >= 2 && _currentGameState.Status == "waiting")
                {
                    Console.WriteLine("Enough players to start the game. Sending start game message...");
                    _ = StartGameAsync();
                }

                HandleGameState(_currentGameState);
                break;

            case MessageType.Error:
                Console.WriteLine($"Error: {message.Data}");
                break;

            default:
                Console.WriteLine($"Unhandled message type: {message.Type}");
                break;
        }
    }

    private void HandleGameState(Game gameState)
    {
        Console.WriteLine($"\nGame State Update:");
        Console.WriteLine($"Game Status: {gameState.Status}");
        Console.WriteLine($"Total Players: {gameState.Players.Count}");
        Console.WriteLine($"Current Turn: {gameState.CurrentTurn}");
        Console.WriteLine($"Current Bet: {gameState.CurrentBet}");
        Console.WriteLine($"Pot: {gameState.Pot}");
        Console.WriteLine($"Dealer Position: {gameState.DealerPosition}");

        var currentPlayer = gameState.Players.FirstOrDefault(p => p.Id == _playerId);
        if (currentPlayer == null)
        {
            Console.WriteLine($"Warning: Could not find current player (ID: {_playerId}) in the game");
            return;
        }

        Console.WriteLine($"\nMy Player Info:");
        Console.WriteLine($"Position: {currentPlayer.Position}");
        Console.WriteLine($"Active: {currentPlayer.Active}");
        Console.WriteLine($"Folded: {currentPlayer.Folded}");
        Console.WriteLine($"Current Bet: {currentPlayer.Bet}");
        Console.WriteLine($"Chips: {currentPlayer.Chips}");

        if (gameState.Status == "started" && gameState.CurrentTurn == currentPlayer.Position)
        {
            var actionNeeded = gameState.CurrentBet > currentPlayer.Bet ? "call" : "check";
            Console.WriteLine($"\nIt's my turn! Taking action: {actionNeeded}");
            Console.WriteLine($"Current bet: {gameState.CurrentBet}, My bet: {currentPlayer.Bet}");
            _ = SendGameActionAsync(actionNeeded);
        }
        else
        {
            Console.WriteLine($"\nNot my turn. Current turn position: {gameState.CurrentTurn}, My position: {currentPlayer.Position}");
        }
    }

    public async Task StartGameAsync()
    {
        if (string.IsNullOrEmpty(_currentGameId))
        {
            Console.WriteLine("No active game to start.");
            return;
        }

        var message = new Message
        {
            Type = MessageType.StartGame,
            GameId = _currentGameId,
            RoomId = _currentRoomId,
            PlayerId = _playerId,
            Data = new { }
        };

        await SendMessageAsync(message);
    }

    public void Dispose()
    {
        _client?.Dispose();
    }
}