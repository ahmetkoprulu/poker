using Bot.Models;
using Websocket.Models;

namespace Bot.Services;

public interface IPokerGameService
{
    Task<GameAction> DetermineNextAction(Game game, string playerId);
    bool IsPlayerTurn(Game game, string playerId);
    int CalculateMinRaise(Game game);
    bool ShouldFold(List<Card> holeCards);
}

public class PokerGameService : IPokerGameService
{
    private readonly Random _random = new Random();

    public async Task<GameAction> DetermineNextAction(Game game, string playerId)
    {
        if (game?.Players == null)
            return null;

        var player = game.Players.FirstOrDefault(p => p.Id == playerId);
        if (player == null || !player.Active || player.Folded)
            return null;

        // Get player's hole cards and community cards
        var holeCards = player.Cards ?? new List<Card>();
        var communityCards = game.CommunityCards ?? new List<Card>();

        // Basic strategy:
        // 1. If we have strong hole cards, be more aggressive
        // 2. If we're in late position, we can play more hands
        // 3. Consider pot odds and implied odds
        // 4. Account for position and number of players still in hand

        // For now, implementing a simple strategy
        if (ShouldFold(holeCards) && game.CurrentBet > 0)
        {
            return new GameAction
            {
                PlayerId = playerId,
                Action = PlayerAction.Fold
            };
        }

        // If no bet and we're in position, we might want to raise
        if (game.CurrentBet == 0)
        {
            if (_random.Next(100) < 30) // 30% chance to raise
            {
                var raiseAmount = CalculateMinRaise(game);
                return new GameAction
                {
                    PlayerId = playerId,
                    Action = PlayerAction.Raise,
                    Amount = raiseAmount
                };
            }

            return new GameAction
            {
                PlayerId = playerId,
                Action = PlayerAction.Check
            };
        }

        // If there's a bet, decide between call or fold
        var handStrength = CalculateHandStrength(holeCards, communityCards);
        var callAmount = game.CurrentBet - player.Bet;
        var potOdds = (double)callAmount / (game.Pot + callAmount);

        if (handStrength > potOdds)
        {
            return new GameAction
            {
                PlayerId = playerId,
                Action = PlayerAction.Call
            };
        }

        return new GameAction
        {
            PlayerId = playerId,
            Action = PlayerAction.Fold
        };
    }

    public async Task<PokerHand> EvaluateHand(List<Card> playerCards, List<Card> communityCards)
    {
        // TODO: Implement proper hand evaluation
        // For now, returning a basic high card evaluation
        return new PokerHand
        {
            Cards = playerCards.Concat(communityCards).ToList(),
            Rank = HandRank.HighCard,
            Value = 0
        };
    }

    public bool IsPlayerTurn(Game game, string playerId)
    {
        if (game?.Players == null || string.IsNullOrEmpty(playerId))
            return false;

        var currentPlayer = game.Players.FirstOrDefault(p => p.Position == game.CurrentTurn);
        return currentPlayer?.Id == playerId && currentPlayer.Active && !currentPlayer.Folded;
    }

    public int CalculateMinRaise(Game game)
    {
        // Minimum raise is current bet + the difference between current bet and previous bet
        // For simplicity, using minBet as the minimum raise size
        return game.CurrentBet + game.MinBet;
    }

    public bool ShouldFold(List<Card> holeCards)
    {
        if (holeCards == null || holeCards.Count != 2)
            return true;

        // Very basic hole card evaluation
        // TODO: Implement proper preflop hand strength calculation
        var card1Value = GetCardValue(holeCards[0].Value);
        var card2Value = GetCardValue(holeCards[1].Value);
        var suited = holeCards[0].Suit == holeCards[1].Suit;

        // Play pairs 77 and higher
        if (card1Value == card2Value && card1Value >= 7)
            return false;

        // Play AK, AQ, AJ
        if (card1Value == 14 || card2Value == 14)
        {
            var otherCard = card1Value == 14 ? card2Value : card1Value;
            if (otherCard >= 11)
                return false;
        }

        // Play suited connectors 89+ if suited
        if (suited && Math.Abs(card1Value - card2Value) == 1 && Math.Min(card1Value, card2Value) >= 8)
            return false;

        return true;
    }

    public int CalculateHandStrength(List<Card> holeCards, List<Card> communityCards)
    {
        // TODO: Implement proper hand strength calculation
        // For now, returning a simple value between 0 and 100
        if (holeCards == null || holeCards.Count != 2)
            return 0;

        var strength = 0;
        var card1Value = GetCardValue(holeCards[0].Value);
        var card2Value = GetCardValue(holeCards[1].Value);
        var suited = holeCards[0].Suit == holeCards[1].Suit;

        // Pair
        if (card1Value == card2Value)
            strength += 60;

        // High cards
        if (card1Value >= 10)
            strength += 10;
        if (card2Value >= 10)
            strength += 10;

        // Suited
        if (suited)
            strength += 10;

        // Connected
        if (Math.Abs(card1Value - card2Value) == 1)
            strength += 10;

        return Math.Min(strength, 100);
    }

    private int GetCardValue(string value)
    {
        return value switch
        {
            "A" => 14,
            "K" => 13,
            "Q" => 12,
            "J" => 11,
            _ => int.Parse(value)
        };
    }
}