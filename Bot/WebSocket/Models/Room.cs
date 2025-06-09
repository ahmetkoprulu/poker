using Bot.Models;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace Websocket.Models;

public class Room
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("status")] public string Status { get; set; }
    [JsonProperty("game")] public Game Game { get; set; }
    [JsonProperty("maxPlayers")] public int MaxPlayers { get; set; }
    [JsonProperty("minBet")] public int MinBet { get; set; }
}

public class RoomSummary
{
    [JsonProperty("id")] public string Id { get; set; }

    [JsonProperty("status")] public string Status { get; set; }

    [JsonProperty("max_room_players")] public int MaxRoomPlayers { get; set; }

    [JsonProperty("players_in_room")] public int PlayersInRoom { get; set; }

    [JsonProperty("game_status")] public string GameStatus { get; set; }

    [JsonProperty("game_type")] public GameType GameType { get; set; }

    [JsonProperty("min_bet")] public int MinBet { get; set; }

    [JsonProperty("max_game_players")] public int MaxGamePlayers { get; set; }

    [JsonProperty("players_in_game")] public int PlayersInGame { get; set; }
}

public class RoomState
{
    [JsonProperty("room_id")] public string RoomId { get; set; }
    [JsonProperty("status")] public string Status { get; set; }
    [JsonProperty("players")] public List<Player> Players { get; set; }
    [JsonProperty("max_players")] public int MaxPlayers { get; set; }
    [JsonProperty("max_game_players")] public int MaxGamePlayers { get; set; }
    [JsonProperty("min_bet")] public int MinBet { get; set; }
    [JsonProperty("game_type")] public GameType GameType { get; set; }
    [JsonProperty("game_status")] public string GameStatus { get; set; }
    [JsonProperty("game_state")] public JToken GameState { get; set; }
}