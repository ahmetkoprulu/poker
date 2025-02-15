using System.Text.Json.Serialization;

namespace Bot.Models;

public static class GameStatus
{
    public const string Waiting = "waiting";
    public const string Started = "started";
    public const string Finished = "finished";
}

public static class Round
{
    public const string PreFlop = "preflop";
    public const string Flop = "flop";
    public const string Turn = "turn";
    public const string River = "river";
}

public static class Suit
{
    public const string Hearts = "hearts";
    public const string Diamonds = "diamonds";
    public const string Clubs = "clubs";
    public const string Spades = "spades";
}

public static class CardValue
{
    public const string Two = "2";
    public const string Three = "3";
    public const string Four = "4";
    public const string Five = "5";
    public const string Six = "6";
    public const string Seven = "7";
    public const string Eight = "8";
    public const string Nine = "9";
    public const string Ten = "10";
    public const string Jack = "J";
    public const string Queen = "Q";
    public const string King = "K";
    public const string Ace = "A";
}

public static class HandRank
{
    public const int HighCard = 0;
    public const int OnePair = 1;
    public const int TwoPair = 2;
    public const int ThreeOfAKind = 3;
    public const int Straight = 4;
    public const int Flush = 5;
    public const int FullHouse = 6;
    public const int FourOfAKind = 7;
    public const int StraightFlush = 8;
    public const int RoyalFlush = 9;
}

public static class PlayerAction
{
    public const string Fold = "fold";
    public const string Check = "check";
    public const string Call = "call";
    public const string Raise = "raise";
}

public class GameAction
{
    [JsonPropertyName("playerId")] public string PlayerId { get; set; }
    [JsonPropertyName("action")] public string Action { get; set; }
    [JsonPropertyName("amount")] public int Amount { get; set; }
}

public class PokerHand
{
    [JsonPropertyName("cards")] public List<Card> Cards { get; set; }
    [JsonPropertyName("rank")] public int Rank { get; set; }
    [JsonPropertyName("value")] public int Value { get; set; }
}

public class BettingRound
{
    [JsonPropertyName("round")] public string Round { get; set; }
    [JsonPropertyName("currentBet")] public int CurrentBet { get; set; }
    [JsonPropertyName("pot")] public int Pot { get; set; }
    [JsonPropertyName("actions")] public List<GameAction> Actions { get; set; }
}