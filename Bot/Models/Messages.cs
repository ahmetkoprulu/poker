using System.Text.Json.Serialization;
using Newtonsoft.Json;

namespace Bot.Models;

public static class MessageType
{
    public const string JoinRoom = "room_join";
    public const string LeaveRoom = "room_leave";
    public const string JoinGame = "game_join";
    public const string LeaveGame = "game_leave";
    public const string StartGame = "game_start";
    public const string GameAction = "game_action";
    public const string GameInfo = "game_info";
    public const string PlayerList = "player_list";
    public const string Error = "error";
}

public class Message
{
    [JsonPropertyName("type")] public string Type { get; set; }
    [JsonPropertyName("data")] public object Data { get; set; }
}

public class Response
{
    [JsonProperty("type")] public string Type { get; set; }
    [JsonProperty("data")] public object Data { get; set; }
}

public class MessageRoomInfo
{
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
}

public class MessageJoinRoom
{
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
}

public class MessageLeaveRoom
{
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
}

public class MessageJoinGame
{
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
    [JsonPropertyName("position")] public int Position { get; set; }
}

public class MessageLeaveGame
{
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
}

public class MessageStartGame
{
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
}

public class MessageGameAction
{
    [JsonPropertyName("action")] public string Action { get; set; }
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
    [JsonPropertyName("data")] public object Data { get; set; }
}
