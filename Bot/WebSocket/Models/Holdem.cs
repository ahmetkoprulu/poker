using System.Text.Json.Serialization;
using Newtonsoft.Json;

namespace Websocket.Models;

public class HoldemState
{
    [JsonProperty("players")] public IEnumerable<HoldemPlayer> Players { get; set; }
    [JsonProperty("community_cards")] public IEnumerable<Card> CommunityCards { get; set; }
    [JsonProperty("pot")] public int Pot { get; set; }
    [JsonProperty("current_bet")] public int CurrentBet { get; set; }
    [JsonProperty("current_round")] public HoldemRound CurrentRound { get; set; }
}

public class HoldemPlayer
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("status")] public PlayerStatus Status { get; set; }
    [JsonProperty("position")] public int Position { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("balance")] public int Balance { get; set; }
    [JsonProperty("hand")] public List<Card> Hand { get; set; }
    [JsonProperty("is_folded")] public bool IsFolded { get; set; }
    [JsonProperty("is_all_in")] public bool IsAllIn { get; set; }
    [JsonProperty("is_dealer")] public bool IsDealer { get; set; }
    [JsonProperty("is_small_blind")] public bool IsSmallBlind { get; set; }
    [JsonProperty("is_big_blind")] public bool IsBigBlind { get; set; }
    [JsonProperty("is_current_turn")] public bool IsCurrentTurn { get; set; }
}

public class HandResult
{
    [JsonProperty("rank")] public HoldemHandRank Rank { get; set; }
    [JsonProperty("high_cards")] public List<int> HighCards { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }
}

public class HoldemMessage
{
    public HoldemMessageType Type { get; set; }
    public object Data { get; set; }
    public long Timestamp { get; set; }
}

public class HoldemActionMessage
{
    public string PlayerId { get; set; }
    public HoldemAction Action { get; set; }
    public int Amount { get; set; }
}

public class HoldemWinnerMessage
{
    public string PlayerId { get; set; }
    public int Amount { get; set; }
}

public class HoldemShowdownMessage
{
    public List<HandResult> Winners { get; set; }
    public int Pot { get; set; }
    public object GameState { get; set; }
}

public class HoldemPlayerTurnMessage
{
    public string PlayerId { get; set; }
    public int Timeout { get; set; }
}

public enum HoldemMessageType : short
{
    GameStart,
    GameEnd,
    RoundStart,
    RoundEnd,
    PlayerTurn,
    PlayerAction,
    Showdown,
    Winner
}

public enum HoldemRound : short
{
    PreFlop,
    Flop,
    Turn,
    River,
    Showdown
}

public enum HoldemHandRank : short
{
    HighCard,
    Pair,
    TwoPair,
    ThreeOfAKind,
    Straight,
    Flush,
    FullHouse,
    FourOfAKind,
    StraightFlush,
    RoyalFlush
}

public enum HoldemAction : short
{
    Fold,
    Check,
    Call,
    Bet,
    Raise,
    AllIn
}