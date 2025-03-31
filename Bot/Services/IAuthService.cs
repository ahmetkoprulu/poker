using Bot.Models;
using Websocket.Models;

namespace Bot.Services;

public interface IAuthService
{
    Task<LoginResponse> LoginAsync(string email, string password);
    Task<LoginResponse> LoginAsGuestAsync();
    Task<Player> GetPlayerAsync(string token);
}