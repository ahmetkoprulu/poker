using Bot.Models;

namespace Bot.Services;

public interface IWebSocketClientService : IDisposable
{
    Task StartAsync();
    Task JoinGameAsync(string roomId, int position = 2);
    Task LeaveGameAsync(string gameId);
    Task SendGameActionAsync(GameAction action);
    Task<Game> GetGameStateAsync(string gameId);
    event Func<Game, Task> OnGameStateChanged;
    event Func<string, Task> OnError;
}