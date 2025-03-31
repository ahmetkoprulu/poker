using System.Drawing;
using Bot.Models;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;
using Pastel;
using Websocket.Models;
using Websocket.Services;

namespace Bot.Services;

public class BotService(IAuthService authService, IWebSocketClient webSocketClient, IConfiguration configuration, ILogger<BotService> logger)
{
    private string _token;
    private string _playerId;
    private IEnumerable<RoomSummary> _rooms;
    private string _currentRoomId;
    private RoomState _currentRoomState;

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

        // _rooms = await GetRooms();
        // if (!_rooms.Any()) return;

        // await JoinRoom(_rooms.First().Id);
    }

    public async Task<IEnumerable<RoomSummary>> GetRooms()
    {
        var wsUrl = configuration["WebSocketUrl"] ?? "ws://localhost:8080/ws";
        await webSocketClient.StartAsync(_playerId, _token, wsUrl);
        var rooms = await webSocketClient.GetRoomsByGameTypeAsync(GameType.Holdem);

        Console.WriteLine("Rooms:".Pastel(Color.LightBlue));
        foreach (var room in rooms)
        {
            var str = $"\t{room.Id} - {room.PlayersInGame} / {room.MaxGamePlayers} ({room.GameStatus})\n\t\tGame Type: {room.GameType} - Min Bet: {room.MinBet}";
            Console.WriteLine(str.Pastel(Color.LightBlue));
        }

        return rooms;
    }

    public async Task JoinRoom(string roomId)
    {
        await webSocketClient.JoinRoomAsync(roomId);
    }

    public void RegisterHandlers()
    {
        var messageHandlerRegistry = new MessageHandlerRegistry();
        messageHandlerRegistry.On<RoomState>(MessageType.RoomJoinOk, message =>
        {
            _currentRoomId = message.RoomId;
            _currentRoomState = message;
            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<Player>(MessageType.RoomJoin, message =>
        {
            Console.WriteLine($"Joined game {message.Id}".Pastel(Color.LightGreen));
            _currentRoomState.Players.Add(message);
            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<Player>(MessageType.RoomLeave, message =>
        {
            Console.WriteLine($"Left game {message.Id}".Pastel(Color.Pink));
            _currentRoomState.Players.Remove(message);
            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<RoomState>(MessageType.GameJoin, message =>
        {
            Console.WriteLine($"Joined game {message.RoomId}".Pastel(Color.LightGreen));
            _currentRoomId = message.RoomId;
            _currentRoomState = message;
            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<MessageGameLeave>(MessageType.GameLeave, message =>
        {
            Console.WriteLine($"Left game {message.RoomId}".Pastel(Color.Pink));

            return Task.CompletedTask;
        });

        messageHandlerRegistry.On<MessageGameAction<HoldemMessage>>(MessageType.GameHoldemAction, message =>
        {
            if (message.RoomId != _currentRoomId) return Task.CompletedTask;
            var data = JsonConvert.SerializeObject(message.Data);

            return Task.CompletedTask;
        });
    }
}