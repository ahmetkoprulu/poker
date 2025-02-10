namespace Bot.Services;

public interface IWebSocketClientService : IDisposable
{
    Task StartAsync();
    Task JoinGameAsync(string roomId);
    Task SendGameActionAsync(string action);
}