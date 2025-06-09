using Newtonsoft.Json;

namespace Websocket.Models;

public class Game
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("status")] public string Status { get; set; }
    [JsonProperty("gameType")] public string GameType { get; set; } // GameType
    [JsonProperty("players")] public List<GamePlayer> Players { get; set; } = new();
    [JsonProperty("minBet")] public int MinBet { get; set; }
    [JsonProperty("maxPlayers")] public int MaxPlayers { get; set; }
}

public class GamePlayer
{
    [JsonProperty("position")] public int Position { get; set; }
    [JsonProperty("balance")] public int Balance { get; set; }
    [JsonProperty("lastAction")] public string LastAction { get; set; }
    [JsonProperty("status")] public string Status { get; set; }
}

public class Card
{
    [JsonProperty("suit")] public string Suit { get; set; }
    [JsonProperty("value")] public string Value { get; set; }
    [JsonProperty("hidden")] public bool Hidden { get; set; }
}

public class GameMessage
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
    [JsonProperty("game_type")] public GameType GameType { get; set; }
    [JsonProperty("data")] public object Data { get; set; }
}

public enum PlayerStatus
{
    Waiting,
    Active,
    Inactive,
}

public static class GameError
{
    public const string GameFull = "game_full";
    public const string GamePlayerAlreadyIn = "game_player_already_in";
    public const string GamePositionTaken = "game_position_taken";
    public const string GamePlayerNotFound = "game_player_not_found";
    public const string GameNotReady = "game_not_ready";
}

public static class GameStatus
{
    public const string Waiting = "waiting";
    public const string Starting = "starting";
    public const string Started = "started";
    public const string Ending = "ending";
    public const string End = "end";
}

public static class GameActionType
{
    public const string PlayerJoin = "player_join";
    public const string PlayerLeave = "player_leave";
    public const string PlayerAction = "player_action";
}

public enum GameType : short
{
    Holdem = 1,
}

public static class GameMessageType
{
    public const string GameStart = "game_start";
    public const string GameEnd = "game_end";
    public const string GamePlayerJoin = "game_player_join";
    public const string GamePlayerLeave = "game_player_leave";
}
