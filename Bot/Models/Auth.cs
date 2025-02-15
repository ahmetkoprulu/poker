using System.Text.Json.Serialization;
using Newtonsoft.Json;

namespace Bot.Models;

public class ApiResponse<T>
{
    [JsonProperty("success")] public bool Success { get; set; }
    [JsonProperty("status")] public int Status { get; set; }
    [JsonProperty("data")] public T Data { get; set; }
    [JsonProperty("message")] public string Message { get; set; }
}

public class LoginRequest
{
    [JsonPropertyName("email")] public string Email { get; set; }
    [JsonPropertyName("password")] public string Password { get; set; }
}

public class LoginResponse
{
    [JsonPropertyName("token")] public string Token { get; set; }
    [JsonPropertyName("user")] public Player Player { get; set; }
}

public class User
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("player")] public Player Player { get; set; }
}

public class Player
{
    [JsonPropertyName("id")] public string Id { get; set; }
    [JsonPropertyName("email")] public string Email { get; set; }
    [JsonPropertyName("name")] public string Name { get; set; }
    [JsonPropertyName("chips")] public int Chips { get; set; }
    [JsonPropertyName("createdAt")] public DateTime CreatedAt { get; set; }
    [JsonPropertyName("updatedAt")] public DateTime UpdatedAt { get; set; }
}

public class RegisterRequest
{
    [JsonProperty("email")] public string Email { get; set; }
    [JsonProperty("password")] public string Password { get; set; }
}

public class PlayerInfo
{
    public string Id { get; set; }
    public string Name { get; set; }
}

