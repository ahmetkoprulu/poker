using Bot.Models;

namespace Bot.Services;

public interface IAuthService
{
    Task<LoginResponse> LoginAsync(string email, string password);
    Task<Player> GetPlayerAsync(string token);
}