using System.Text.Json;
using System.Text.Json.Serialization;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace Websocket.Models;

public static class MessageType
{
    public const string RoomInfo = "room_info";
    public const string RoomJoin = "room_join";
    public const string RoomJoinOk = "room_join_ok";
    public const string RoomLeave = "room_leave";
    public const string RoomLeaveOk = "room_leave_ok";

    public const string GameJoin = "game_join";
    public const string GameJoinOk = "game_join_ok";
    public const string GameLeave = "game_leave";
    public const string GameLeaveOk = "game_leave_ok";

    public const string GameAction = "game_action";
    public const string GameHoldemAction = "game_holdem_action";

    public const string GameInfo = "game_info";
    public const string PlayerList = "player_list";
    public const string Error = "error";
}

public class Message
{
    [JsonPropertyName("type")] public string Type { get; set; } // MessageType
    [JsonPropertyName("data")] public JsonElement Data { get; set; }
}

public class Message<T>
{
    [JsonPropertyName("type")] public string Type { get; set; } // MessageType
    [JsonPropertyName("data")] public T Data { get; set; }
}

public class Response
{
    [JsonPropertyName("type")] public string Type { get; set; } // MessageType
    [JsonPropertyName("data")] public JsonElement Data { get; set; }
    [JsonPropertyName("timestamp")] public DateTime Timestamp { get; set; }
}

public class Response<T>
{
    [JsonPropertyName("type")] public string Type { get; set; } // MessageType
    [JsonPropertyName("data")] public T Data { get; set; }
    [JsonPropertyName("timestamp")] public DateTime Timestamp { get; set; }
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

public class MessageRoomJoinResponse
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
    [JsonProperty("state")] public JToken State { get; set; }
}

public class MessageRoomLeave
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageRoomLeaveResponse
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
    [JsonProperty("state")] public JToken State { get; set; }
}

public class MessageGameJoin
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
    [JsonProperty("position")] public int Position { get; set; }
}

public class MessageGameJoinResponse
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player")] public Player Player { get; set; }
    [JsonProperty("position")] public int Position { get; set; }
    [JsonProperty("state")] public JToken State { get; set; }
}

public class MessageGameJoinOk
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageGameJoinOkResponse
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("state")] public JToken State { get; set; }
}

public class MessageGameLeave
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class MessageGameLeaveResponse
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
    [JsonProperty("state")] public JToken State { get; set; }
}

public class MessageGameAction
{
    [JsonPropertyName("room_id")] public string RoomId { get; set; }
    [JsonPropertyName("player_id")] public string PlayerId { get; set; }
    [JsonPropertyName("game_type")] public GameType GameType { get; set; }
    [JsonPropertyName("data")] public JToken Data { get; set; }
}
