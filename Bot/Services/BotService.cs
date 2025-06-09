using System.Drawing;
using Bot.Models;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;
using Pastel;
using Websocket.Models;
using Websocket.Services;

namespace Bot.Services;

public class BotService(IAuthService authService, IWebSocketClient webSocketClient, MessageHandlerRegistry messageHandlerRegistry, IGameSerivice<HoldemGameState, HoldemActionMessage> gameService, IConfiguration configuration, ILogger<BotService> logger)
{
    private string WsUrl { get; set; }
    private string _token;
    private string _playerId;
    private IEnumerable<RoomSummary> _rooms;
    private string? _currentRoomId;
    private RoomState? _currentRoomState;
    private DateTime _lastStateTimestamp = DateTime.MinValue;

    public async Task StartAsync()
    {
        // Authenticate first
        logger.LogInformation("Authenticating as guest");
        var loginResponse = await authService.LoginAsGuestAsync();
        _token = loginResponse.Token;
        _playerId = loginResponse.User.Player.Id;

        // Get player details
        var player = await authService.GetPlayerAsync(_token);
        logger.LogInformation("Authenticated as player {PlayerName} with {Chips} chips", player.Username, player.Chips);

        WsUrl = configuration["WebSocketUrl"] ?? "ws://localhost:8080";
        await webSocketClient.StartAsync(_playerId, _token, WsUrl);
        RegisterHandlers();

        _rooms = await GetRooms();
        if (!_rooms.Any()) return;

        await JoinRoom(_rooms.First().Id);
    }

    public async Task<IEnumerable<RoomSummary>> GetRooms()
    {
        var rooms = await webSocketClient.GetRoomsByGameTypeAsync(GameType.Holdem);

        Console.WriteLine("Rooms:".Pastel(Color.YellowGreen));
        foreach (var room in rooms)
        {
            Console.WriteLine($"\t{room.Id} ({room.GameType}) - {room.PlayersInGame} / {room.MaxGamePlayers} ({room.GameStatus}) - Min Bet: {room.MinBet}\n".Pastel(Color.YellowGreen));
        }

        return rooms;
    }

    public async Task JoinRoom(string roomId)
    {
        await webSocketClient.JoinRoomAsync(roomId);
    }

    public async Task TryJoinGame(string roomId, int? position = null)
    {
        position ??= gameService.GetAvailablePosition();
        if (position == null)
        {
            Console.WriteLine("No available position found".Pastel(Color.IndianRed));
            return;
        }

        await webSocketClient.JoinGameAsync(roomId, position.Value);
    }

    public void SetState(RoomState state, DateTime timestamp)
    {
        if (timestamp <= _lastStateTimestamp)
        {
            Console.WriteLine("Skipping state update".Pastel(Color.Gray));
            return;
        }

        _lastStateTimestamp = timestamp;
        _currentRoomState = state;
        gameService.SetState(state.GameState);
        Console.WriteLine($"State is updated".Pastel(Color.LightGreen));
    }

    public void RegisterHandlers()
    {
        // Handles when the room join message is successfull
        messageHandlerRegistry.On<RoomState>(MessageType.RoomJoinOk, (message, timestamp) =>
        {
            Console.WriteLine($"You joined Room {message.RoomId} - Type: {message.GameType}, Game Status: {message.GameStatus}".Pastel(Color.YellowGreen));
            Console.WriteLine("Players: ".Pastel(Color.YellowGreen));
            foreach (var player in message.Players)
            {
                Console.WriteLine($"\t{player.Id} {player.Username}".Pastel(Color.YellowGreen));
            }

            SetState(message, timestamp);
            _ = TryJoinGame(_currentRoomId);

            return Task.CompletedTask;
        });

        // Handles the new player joins the room.
        messageHandlerRegistry.On<MessageRoomJoinResponse>(MessageType.RoomJoin, (message, timestamp) =>
        {
            Console.WriteLine($"The Player {message.PlayerId} Joined Room {message.RoomId}".Pastel(Color.YellowGreen));

            SetState(message.State.ToObject<RoomState>(), timestamp);
            return Task.CompletedTask;
        });


        // Handles the room leave message is successfull
        messageHandlerRegistry.On<RoomState>(MessageType.RoomLeaveOk, (message, timestamp) =>
        {
            Console.WriteLine($"You left room {_currentRoomId}".Pastel(Color.Pink));
            _currentRoomId = null;
            _currentRoomState = null;
            return Task.CompletedTask;
        });

        // Handles the a player lefts the room
        messageHandlerRegistry.On<MessageRoomLeaveResponse>(MessageType.RoomLeave, (message, timestamp) =>
        {
            Console.WriteLine($"The Player {message.PlayerId} left room {message.RoomId}".Pastel(Color.Pink));
            var state = message.State.ToObject<RoomState>();
            SetState(state, timestamp);

            return Task.CompletedTask;
        });

        // Handles the game join request is successfull
        messageHandlerRegistry.On<MessageGameJoinResponse>(MessageType.GameJoinOk, (message, timestamp) =>
        {
            Console.WriteLine($"You joined game {message.RoomId} the game at position {message.Position}.".Pastel(Color.LightGreen));
            var state = message.State.ToObject<RoomState>();
            SetState(state, timestamp);

            return Task.CompletedTask;
        });

        // Handles the new player joins the game
        messageHandlerRegistry.On<MessageGameJoinResponse>(MessageType.GameJoin, (message, timestamp) =>
        {
            Console.WriteLine($"The Player {message.Player.Id} joined game at position {message.Position} in {message.RoomId}.".Pastel(Color.LightGreen));
            var state = message.State.ToObject<RoomState>();
            SetState(state, timestamp);

            return Task.CompletedTask;
        });


        messageHandlerRegistry.On<MessageGameLeaveResponse>(MessageType.GameLeave, (message, timestamp) =>
        {
            Console.WriteLine($"You left game {message.RoomId}.".Pastel(Color.Pink));
            var state = message.State.ToObject<RoomState>();
            SetState(state, timestamp);

            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<MessageGameLeave>(MessageType.GameLeave, (message, timestamp) =>
        {
            Console.WriteLine($"Left game {message.RoomId}".Pastel(Color.Pink));
            // var state = message.State.ToObject<RoomState>();
            // SetState(state, timestamp);

            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<MessageGameAction>(MessageType.GameHoldemAction, (message, timestamp) =>
        {
            if (message.RoomId != _currentRoomId) return Task.CompletedTask;
            var data = JsonConvert.SerializeObject(message.Data);

            return Task.CompletedTask;
        });
    }
}