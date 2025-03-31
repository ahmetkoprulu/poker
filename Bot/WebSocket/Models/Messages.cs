using System.Text.Json.Serialization;
using Newtonsoft.Json;

namespace Websocket.Models;

public static class MessageType
{
    public const string RoomInfo = "room_info";
    public const string RoomJoin = "room_join";
    public const string RoomJoinOk = "room_join_ok";
    public const string RoomLeave = "room_leave";

    public const string GameJoin = "game_join";
    public const string GameJoinOk = "game_join_ok";
    public const string GameLeave = "game_leave";
    public const string GameAction = "game_action";
    public const string GameHoldemAction = "game_holdem_action";
    public const string GameInfo = "game_info";
    public const string PlayerList = "player_list";
    public const string Error = "error";
}

public class Message<T>
{
    [JsonPropertyName("type")] public string Type { get; set; } // MessageType
    [JsonPropertyName("data")] public T Data { get; set; }
}

public class Error
{
    [JsonProperty("type")] public string Type { get; set; } // MessageType
    [JsonProperty("message")] public string Message { get; set; }
}

public class MessageRoomInfo
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
}

public class MessageRoomJoin
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageRoomLeave
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageGameJoin
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
    [JsonProperty("position")] public int Position { get; set; }
}

public class MessageGameJoinOk
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageGameLeave
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageGameAction<T>
{
    [JsonPropertyName("room_id")] public string RoomId { get; set; }
    [JsonPropertyName("player_id")] public string PlayerId { get; set; }
    [JsonPropertyName("game_type")] public GameType GameType { get; set; }
    [JsonPropertyName("data")] public T Data { get; set; }
}
