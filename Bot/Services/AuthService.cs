using System.Net.Http.Json;
using System.Text.Json.Serialization;
using Bot.Config;
using Bot.Models;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Websocket.Models;

namespace Bot.Services;

public class AuthService(IHttpClientFactory httpClientFactory, IConfiguration configuration) : IAuthService
{
    private readonly string _baseUrl = configuration["ApiBaseUrl"] ?? throw new Exception("ApiBaseUrl is not configured");
    private string? _token;

    public async Task<LoginResponse> LoginAsync(string email, string password)
    {
        var client = httpClientFactory.CreateClient();
        client.BaseAddress = new Uri(_baseUrl);

        var loginResponse = await client.PostAsJsonAsync("/api/v1/auth/login", new
        {
            provider = 1,
            identifier = email,
            secret = password
        });

        if (!loginResponse.IsSuccessStatusCode)
        {
            var error = await loginResponse.Content.ReadAsStringAsync();
            throw new Exception($"Login failed: {error}");
        }

        var loginResult = await loginResponse.Content.ReadFromJsonAsync<ApiResponse<LoginResponse>>() ?? throw new Exception("Login result is null");
        _token = loginResult?.Data?.Token ?? throw new Exception("Token is Empty");

        return loginResult.Data;
    }

    public async Task<LoginResponse> LoginAsGuestAsync()
    {
        var client = httpClientFactory.CreateClient();
        client.BaseAddress = new Uri(_baseUrl);

        var loginResponse = await client.PostAsJsonAsync("/api/v1/auth/login", new
        {
            provider = 0,
            identifier = Guid.NewGuid().ToString(),
        });

        if (!loginResponse.IsSuccessStatusCode)
        {
            var error = await loginResponse.Content.ReadAsStringAsync();
            throw new Exception($"Login failed: {error}");
        }

        var loginResult = await loginResponse.Content.ReadFromJsonAsync<ApiResponse<LoginResponse>>() ?? throw new Exception("Login result is null");
        _token = loginResult?.Data?.Token ?? throw new Exception("Token is Empty");

        return loginResult.Data;
    }

    public async Task<Player> GetPlayerAsync(string token)
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

        var result = await playerResponse.Content.ReadFromJsonAsync<ApiResponse<Player>>() ?? throw new Exception("Player info is null");
        return result.Data;
    }
}