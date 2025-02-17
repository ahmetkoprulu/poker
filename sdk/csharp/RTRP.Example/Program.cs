using System.Net.Http.Headers;
using Microsoft.Extensions.DependencyInjection;
using RTRP.SDK;
using RTRP.SDK.Models;

class Program
{
    private static Client _client = null!;
    private static HttpClient _httpClient = null!;
    private const string BaseUrl = "http://localhost:8080/api/v1";

    static async Task Main(string[] args)
    {
        // Setup dependency injection
        var services = new ServiceCollection();
        ConfigureServices(services);
        var serviceProvider = services.BuildServiceProvider();

        // Get the HTTP client from DI
        _httpClient = serviceProvider.GetRequiredService<HttpClient>();

        // Initialize API client
        _client = new Client(_httpClient);

        try
        {
            // Register a new user
            await RegisterUser();

            // Login
            var loginResult = await Login();
            Console.WriteLine($"Logged in successfully. Player ID: {loginResult.Data.Player.Id}");

            // Set the authorization token for subsequent requests
            _httpClient.DefaultRequestHeaders.Authorization =
                new AuthenticationHeaderValue("Bearer", loginResult.Data.Id);

            // Get player profile
            var player = await GetPlayerProfile();
            Console.WriteLine($"Player: {player.Username}, Chips: {player.Chips}");

            // List all events
            var events = await ListEvents();
            Console.WriteLine($"Found {events.Count} events");

            if (events.Any())
            {
                var firstEvent = events.First();
                Console.WriteLine($"First event: {firstEvent.Name}");

                // Get event schedules
                var schedules = await GetEventSchedules(firstEvent.Id);
                Console.WriteLine($"Event has {schedules.Count} schedules");

                if (schedules.Any())
                {
                    var firstSchedule = schedules.First();

                    // Get or create player event state
                    var playerEvent = await GetPlayerEventState(firstSchedule.Id);
                    Console.WriteLine($"Player event state: Tickets: {playerEvent.Tickets_left}, Score: {playerEvent.Score}");

                    // Play the event
                    if (playerEvent.Tickets_left > 0)
                    {
                        var playResult = await PlayEvent(firstSchedule.Id);
                        Console.WriteLine($"Play result - Score: {playResult.Score}, Rewards: {playResult.Rewards.Count}");
                        Console.WriteLine($"Score: {playResult.Player_event.Score}");
                    }
                }
            }
        }
        catch (ApiException ex)
        {
            Console.WriteLine($"API Error: {ex.Message}");
            if (ex.Response != null)
            {
                Console.WriteLine($"Status Code: {ex.StatusCode}");
                Console.WriteLine($"Response: {ex.Response}");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error: {ex.Message}");
        }
    }

    private static void ConfigureServices(IServiceCollection services)
    {
        services.AddHttpClient();
    }

    private static async Task RegisterUser()
    {
        var request = new RegisterRequest
        {
            Provider = SocialNetwork._1,
            Identifier = "test@example.com",
            Secret = "password123"
        };

        var result = await _client.RegisterAsync(request);
        Console.WriteLine($"Registration successful: {result.Data}");
    }

    private static async Task<ApiResponseModels_UserPlayer> Login()
    {
        var request = new LoginRequest
        {
            Provider = SocialNetwork._1,
            Identifier = "test@example.com",
            Secret = "password123"
        };

        return await _client.LoginAsync(request);
    }

    private static async Task<Player> GetPlayerProfile()
    {
        return await _client.MeAsync();
    }

    private static async Task<ICollection<Event>> ListEvents()
    {
        return await _client.EventsAllAsync();
    }

    private static async Task<ICollection<EventSchedule>> GetEventSchedules(string eventId)
    {
        return await _client.SchedulesAllAsync(eventId);
    }

    private static async Task<PlayerEvent> GetPlayerEventState(string scheduleId)
    {
        return await _client.PlayerAsync(scheduleId);
    }

    private static async Task<EventPlayResult> PlayEvent(string scheduleId)
    {
        var playData = new PlayEventRequest
        {
            Play_data = new Dictionary<string, object>
            {
                { "action", "spin" },
                { "bet", 100 }
            }
        };

        return await _client.PlayAsync(scheduleId, playData);
    }
}
