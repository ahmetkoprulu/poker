using System.Text.Json.Serialization;

namespace Websocket.Models;

public class GameAction
{
    [JsonPropertyName("action")]
    public string Action { get; set; }

    [JsonPropertyName("amount")]
    public int? Amount { get; set; }
}