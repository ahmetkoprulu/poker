using System.Net.Http.Json;
using System.Text.Json.Serialization;
using Bot.Config;
using Bot.Models;
using Microsoft.Extensions.DependencyInjection;

namespace Bot.Services;

public class AuthService(IHttpClientFactory httpClientFactory, BotConfig config) : IAuthService
{
    private readonly string _baseUrl = config.ApiBaseUrl;
    private string? _token;

    public async Task<string> AuthenticateAsync(string email, string password)
    {
        var client = httpClientFactory.CreateClient();
        client.BaseAddress = new Uri(_baseUrl);

        var loginResponse = await client.PostAsJsonAsync("/api/v1/auth/login", new
        {
            email,
            password
        });

        if (!loginResponse.IsSuccessStatusCode)
        {
            var error = await loginResponse.Content.ReadAsStringAsync();
            throw new Exception($"Login failed: {error}");
        }

        var loginResult = await loginResponse.Content.ReadFromJsonAsync<ApiResponse<LoginResponse>>() ?? throw new Exception("Login result is null");
        _token = loginResult?.Data?.Token ?? throw new Exception("Token is Empty");

        return _token;
    }

    public async Task<UserPlayer> GetPlayerAsync()
    {
        if (string.IsNullOrEmpty(_token)) throw new Exception("Token is not set");

        var client = httpClientFactory.CreateClient();
        client.BaseAddress = new Uri(_baseUrl);
        client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", _token);

        var playerResponse = await client.GetAsync("/api/v1/players/me");
        if (!playerResponse.IsSuccessStatusCode)
        {
            var error = await playerResponse.Content.ReadAsStringAsync();
            throw new Exception($"Failed to get player info: {error}");
        }

        var result = await playerResponse.Content.ReadFromJsonAsync<ApiResponse<UserPlayer>>() ?? throw new Exception("Player info is null");
        return result.Data;
    }
}