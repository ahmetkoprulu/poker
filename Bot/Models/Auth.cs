using System.Text.Json.Serialization;
using Newtonsoft.Json;
using Websocket.Models;

namespace Bot.Models;

public class ApiResponse<T>
{
    [JsonProperty("success")] public bool Success { get; set; }
    [JsonProperty("status")] public int Status { get; set; }
    [JsonProperty("data")] public T Data { get; set; }
    [JsonProperty("message")] public string Message { get; set; }
}

public class ErrorResponse
{
    [JsonProperty("error")] public string Error { get; set; }
}

public class LoginRequest
{
    [JsonPropertyName("email")] public string Email { get; set; }
    [JsonPropertyName("password")] public string Password { get; set; }
}

public class LoginResponse
{
    [JsonPropertyName("token")] public string Token { get; set; }
    [JsonPropertyName("user")] public User User { get; set; }
}

public class RegisterRequest
{
    [JsonProperty("email")] public string Email { get; set; }
    [JsonProperty("password")] public string Password { get; set; }
}

