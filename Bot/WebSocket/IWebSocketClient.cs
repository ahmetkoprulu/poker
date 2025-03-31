using Websocket.Models;

namespace Websocket.Services;

public interface IWebSocketClient : IDisposable
{
    event Func<string, Task> OnError;
    string PlayerId { get; }
    Task StartAsync(string playerId, string token, string wsUrl);
    Task JoinRoomAsync(string roomId);
    Task LeaveRoomAsync(string roomId);
    Task JoinGameAsync(string roomId, int position);
    Task LeaveGameAsync(string roomId);
    Task SendGameActionAsync<T>(string roomId, T action);
    Task<IEnumerable<RoomSummary>> GetRoomsByGameTypeAsync(GameType gameType);
}