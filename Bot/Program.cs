using Bot.Services;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;

namespace Bot;

public class Program
{
    public static async Task Main(string[] args)
    {
        var host = CreateHostBuilder(args).Build();
        var webSocketClient = host.Services.GetRequiredService<IWebSocketClientService>();
        var logger = host.Services.GetRequiredService<ILogger<Program>>();

        try
        {
            await webSocketClient.StartAsync();
            logger.LogInformation("Bot started successfully");

            // Join a game room
            var roomId = host.Services.GetRequiredService<IConfiguration>()["RoomId"];
            if (!string.IsNullOrEmpty(roomId))
            {
                await webSocketClient.JoinGameAsync(roomId);
                logger.LogInformation("Joined game room: {RoomId}", roomId);
            }

            // Keep the application running
            await host.RunAsync();
        }
        catch (Exception ex)
        {
            logger.LogError(ex, "An error occurred while running the bot");
        }
    }

    private static IHostBuilder CreateHostBuilder(string[] args) =>
        Host.CreateDefaultBuilder(args)
            .ConfigureServices((hostContext, services) =>
            {
                // Configure HTTP client
                services.AddHttpClient<IAuthService, AuthService>();

                // Add services
                services.AddSingleton<IAuthService, AuthService>();
                services.AddSingleton<IPokerGameService, PokerGameService>();
                services.AddSingleton<IWebSocketClientService, WebSocketClientService>();

                // Configure logging
                services.AddLogging(builder =>
                {
                    builder.AddConsole();
                    builder.SetMinimumLevel(LogLevel.Error);
                });
            })
            .ConfigureAppConfiguration((hostContext, config) =>
            {
                config.SetBasePath(Directory.GetCurrentDirectory())
                    .AddJsonFile("appsettings.json", optional: false)
                    .AddJsonFile($"appsettings.{hostContext.HostingEnvironment.EnvironmentName}.json", optional: true)
                    .AddEnvironmentVariables()
                    .AddCommandLine(args);
            });
}

public enum SocialNetwork
{
    Guest,
    Email,
    Google,
    Facebook,
    Apple
}

public class User
{
    [JsonProperty("id")] public string ID { get; set; }
    [JsonProperty("provider")] public SocialNetwork Provider { get; set; }
    [JsonProperty("identifier")] public string Identifier { get; set; }
    [JsonProperty("password")] public string Password { get; set; }
    [JsonProperty("profile")] public Profile Profile { get; set; }
    [JsonProperty("player")] public Player? Player { get; set; }
}

public class Profile
{
    [JsonProperty("email")] public string Email { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("phone")] public string Phone { get; set; }
}

public class Player
{
    [JsonProperty("id")] public string ID { get; set; }
    [JsonProperty("username")] public string Username { get; set; }
    [JsonProperty("profile_pic_url")] public string ProfilePicURL { get; set; }
    [JsonProperty("user_id")] public string UserID { get; set; }
    [JsonProperty("chips")] public int Chips { get; set; }
}

public class LoginResponse
{
    [JsonProperty("user")] public User User { get; set; }
    [JsonProperty("token")] public string Token { get; set; }
}

public class Event
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("type")] public EventType Type { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("assets")] public List<Asset> Assets { get; set; }
    [JsonProperty("config")] public object Config { get; set; }
}

public class PathEventConfig
{
    [JsonProperty("entry_fee")] public int EntryFee { get; set; }
    [JsonProperty("dice_count")] public int DiceCount { get; set; }
    [JsonProperty("dice_sides")] public int DiceSides { get; set; }
    [JsonProperty("steps")] public List<PathStep> Steps { get; set; }
}

public class PathStep
{
    [JsonProperty("reward")] public Item Reward { get; set; }
}

public class Item
{
    [JsonProperty("type")] public ItemType Type { get; set; }
    [JsonProperty("amount")] public int Amount { get; set; }
    [JsonProperty("metadata")] public object Metadata { get; set; }
}

public enum ItemType : short
{
    Chips = 1,
    Gold = 2,
}

public enum EventType : short
{
    Slot = 1,
    Pop = 2,
    Path = 3,
}

public class Asset
{
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("url")] public string URL { get; set; }
    [JsonProperty("type")] public AssetType Type { get; set; }
}

public enum AssetType : short
{
    Image = 1,
    Video = 2,
    Audio = 3
}

public class EventSchedule
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("event_id")] public string EventId { get; set; }
    [JsonProperty("start_time")] public DateTime StartTime { get; set; }
    [JsonProperty("end_time")] public DateTime EndTime { get; set; }
}

public class ActiveEventSchedule
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("event_id")] public string EventId { get; set; }
    [JsonProperty("start_time")] public DateTime StartTime { get; set; }
    [JsonProperty("end_time")] public DateTime EndTime { get; set; }

    [JsonProperty("type")] public EventType Type { get; set; }
    [JsonProperty("name")] public string Name { get; set; }
    [JsonProperty("assets")] public List<Asset> Assets { get; set; }
}

public class PlayerEvent
{
    [JsonProperty("id")] public string Id { get; set; }
    [JsonProperty("schedule_id")] public string ScheduleId { get; set; }
    [JsonProperty("player_id")] public string PlayerId { get; set; }

    [JsonProperty("score")] public int Score { get; set; }
    [JsonProperty("attempts")] public int Attempts { get; set; }
    [JsonProperty("last_play")] public DateTime LastPlay { get; set; }

    [JsonProperty("tickets_left")] public int TicketsLeft { get; set; }
    [JsonProperty("state")] public object State { get; set; }
}

public class PlayerEventSchedule
{
    [JsonProperty("schedule_id")] public string ScheduleId { get; set; }
    [JsonProperty("score")] public int Score { get; set; }
    [JsonProperty("attempts")] public int Attempts { get; set; }
    [JsonProperty("last_play")] public DateTime LastPlay { get; set; }
    [JsonProperty("tickets_left")] public int TicketsLeft { get; set; }
    [JsonProperty("state")] public object State { get; set; }
}

public class PlayEventRequest
{
    [JsonProperty("play_data")] public object PlayData { get; set; }
}

public class PlayEventResponse
{
    [JsonProperty("player_event")] public PlayerEvent PlayerEvent { get; set; }
    [JsonProperty("rewards")] public List<Item> Rewards { get; set; }
    [JsonProperty("data")] public object Data { get; set; }
}

public class PathGameResult
{
    [JsonProperty("total_steps")] public int TotalSteps { get; set; }
    [JsonProperty("last_roll")] public List<int> LastRoll { get; set; }
}

