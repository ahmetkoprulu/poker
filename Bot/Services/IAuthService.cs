using Bot.Models;

namespace Bot.Services;

public interface IAuthService
{
    Task<string> AuthenticateAsync(string email, string password);
    Task<UserPlayer> GetPlayerAsync();
}