using System.Text.Json.Serialization;

namespace Bot.Models;

public class ApiResponse<T>
{
    [JsonPropertyName("success")] public bool Success { get; set; }
    [JsonPropertyName("status")] public int Status { get; set; }
    [JsonPropertyName("data")] public T Data { get; set; }
    [JsonPropertyName("message")] public string Message { get; set; }
}

internal class JsonPropertyAttribute : Attribute
{
}

public class LoginResponse
{
    [JsonPropertyName("token")] public string? Token { get; set; }
    [JsonPropertyName("user")] public User? User { get; set; }
}

public class PlayerInfo
{
    public string Id { get; set; }
    public string Name { get; set; }
}

public class User
{
    [JsonPropertyName("id")] public string Id { get; set; }
    [JsonPropertyName("player")] public Player Player { get; set; }
}

public class UserPlayer
{
    [JsonPropertyName("id")] public string Id { get; set; }
    [JsonPropertyName("chips")] public long Chips { get; set; }
}