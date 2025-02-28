using Newtonsoft.Json;

namespace Bot.Models;

public class Card
{
    [JsonProperty("suit")] public string Suit { get; set; }
    [JsonProperty("value")] public string Value { get; set; }
    [JsonProperty("hidden")] public bool Hidden { get; set; }
}

public class GamePlayer
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("cards")] public List<Card> Cards { get; set; } = new();
    [JsonProperty("chips")] public int Chips { get; set; }
    [JsonProperty("bet")] public int Bet { get; set; }
    [JsonProperty("position")] public int Position { get; set; }
    [JsonProperty("active")] public bool Active { get; set; }
    [JsonProperty("folded")] public bool Folded { get; set; }
    [JsonProperty("lastAction")] public string LastAction { get; set; }
}

public class Game
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("status")] public string Status { get; set; }
    [JsonProperty("players")] public List<GamePlayer> Players { get; set; } = new();
    [JsonProperty("communityCards")] public List<Card> CommunityCards { get; set; } = new();
    [JsonProperty("pot")] public int Pot { get; set; }
    [JsonProperty("currentBet")] public int CurrentBet { get; set; }
    [JsonProperty("dealerPosition")] public int DealerPosition { get; set; }
    [JsonProperty("currentTurn")] public int CurrentTurn { get; set; }
    [JsonProperty("minBet")] public int MinBet { get; set; }
    [JsonProperty("maxPlayers")] public int MaxPlayers { get; set; }
    [JsonProperty("createdAt")] public DateTime CreatedAt { get; set; }
    [JsonProperty("updatedAt")] public DateTime UpdatedAt { get; set; }
}

public class Room
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("status")] public string Status { get; set; }
    [JsonProperty("game")] public Game Game { get; set; }
    [JsonProperty("maxPlayers")] public int MaxPlayers { get; set; }
    [JsonProperty("minBet")] public int MinBet { get; set; }
    [JsonProperty("createdAt")] public DateTime CreatedAt { get; set; }
    [JsonProperty("updatedAt")] public DateTime UpdatedAt { get; set; }
}