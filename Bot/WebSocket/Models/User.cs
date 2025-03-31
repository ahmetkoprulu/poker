using Newtonsoft.Json;

namespace Websocket.Models;

public class User
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("player")] public Player Player { get; set; }
}

public class Player
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("username")] public string Username { get; set; }
    [JsonProperty("profile_pic_url")] public string ProfilePicURL { get; set; }
    [JsonProperty("chips")] public int Chips { get; set; }
}