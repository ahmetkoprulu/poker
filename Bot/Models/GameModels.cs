using System.Text.Json.Serialization;

namespace Bot.Models;

public static class MessageType
{
    public const string JoinGame = "join_game";
    public const string LeaveGame = "leave_game";
    public const string StartGame = "start_game";
    public const string GameAction = "game_action";
    public const string RoomInfo = "room_info";
    public const string GameState = "game_state";
    public const string PlayerList = "player_list";
    public const string Error = "error";
}

public class Message
{
    [JsonPropertyName("type")] public string Type { get; set; }
    [JsonPropertyName("roomId")] public string RoomId { get; set; }
    [JsonPropertyName("gameId")] public string GameId { get; set; }
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
    [JsonPropertyName("data")] public object Data { get; set; }
}

public class Card
{
    [JsonPropertyName("suit")] public string Suit { get; set; }
    [JsonPropertyName("value")] public string Value { get; set; }
    [JsonPropertyName("hidden")] public bool Hidden { get; set; }
}

public class GamePlayer
{
    [JsonPropertyName("id")] public string Id { get; set; }
    [JsonPropertyName("name")] public string Name { get; set; }
    [JsonPropertyName("cards")] public List<Card> Cards { get; set; } = new();
    [JsonPropertyName("chips")] public int Chips { get; set; }
    [JsonPropertyName("bet")] public int Bet { get; set; }
    [JsonPropertyName("position")] public int Position { get; set; }
    [JsonPropertyName("active")] public bool Active { get; set; }
    [JsonPropertyName("folded")] public bool Folded { get; set; }
    [JsonPropertyName("lastAction")] public string LastAction { get; set; }
}

public class Game
{
    [JsonPropertyName("id")] public string Id { get; set; }
    [JsonPropertyName("status")] public string Status { get; set; }
    [JsonPropertyName("players")] public List<GamePlayer> Players { get; set; } = new();
    [JsonPropertyName("communityCards")] public List<Card> CommunityCards { get; set; } = new();
    [JsonPropertyName("pot")] public int Pot { get; set; }
    [JsonPropertyName("currentBet")] public int CurrentBet { get; set; }
    [JsonPropertyName("dealerPosition")] public int DealerPosition { get; set; }
    [JsonPropertyName("currentTurn")] public int CurrentTurn { get; set; }
    [JsonPropertyName("minBet")] public int MinBet { get; set; }
    [JsonPropertyName("maxPlayers")] public int MaxPlayers { get; set; }
    [JsonPropertyName("createdAt")] public DateTime CreatedAt { get; set; }
    [JsonPropertyName("updatedAt")] public DateTime UpdatedAt { get; set; }
}

public class Room
{
    [JsonPropertyName("id")] public string Id { get; set; }
    [JsonPropertyName("name")] public string Name { get; set; }
    [JsonPropertyName("status")] public string Status { get; set; }
    [JsonPropertyName("game")] public Game Game { get; set; }
    [JsonPropertyName("maxPlayers")] public int MaxPlayers { get; set; }
    [JsonPropertyName("minBet")] public int MinBet { get; set; }
    [JsonPropertyName("createdAt")] public DateTime CreatedAt { get; set; }
    [JsonPropertyName("updatedAt")] public DateTime UpdatedAt { get; set; }
}