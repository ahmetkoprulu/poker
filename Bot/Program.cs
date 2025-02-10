using Bot.Config;
using Bot.Services;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

var host = Host.CreateDefaultBuilder(args)
    .ConfigureServices((context, services) =>
    {
        // Add configuration
        services.AddSingleton<BotConfig>(sp =>
        {
            var config = new ConfigurationBuilder()
                .SetBasePath(Directory.GetCurrentDirectory())
                .AddJsonFile("appsettings.json")
                .Build();

            return config.Get<BotConfig>() ?? throw new Exception("BotConfig is null");
        });

        // Add HttpClient
        services.AddHttpClient();

        // Add services
        services.AddScoped<IAuthService>(sp =>
        {
            var config = sp.GetRequiredService<BotConfig>();
            var httpClientFactory = sp.GetRequiredService<IHttpClientFactory>();
            return new AuthService(httpClientFactory, config);
        });
    })
    .Build();

var botConfig = host.Services.GetRequiredService<BotConfig>();
if (botConfig == null)
{
    Console.WriteLine("Failed to load configuration.");
    return;
}

Console.WriteLine($"Starting authentication process for bot: {botConfig.Email}");

try
{
    Console.WriteLine("Authenticating...");
    // Fetch player info
    var authService = host.Services.GetRequiredService<IAuthService>();
    var token = await authService.AuthenticateAsync(botConfig.Email, botConfig.Password);
    var player = await authService.GetPlayerAsync();

    Console.WriteLine($"Authentication successful. Player ID: {player.Id}, Chips: {player.Chips}");
    Console.WriteLine($"--------------------------------");

    Console.WriteLine($"Connecting to game server: {botConfig.WebSocketUrl}");

    // Initialize WebSocket client with authenticated player info
    using var client = new WebSocketClientService(botConfig.WebSocketUrl, token, player);
    await client.StartAsync();
    Console.WriteLine("Connected to WebSocket server");
    Console.WriteLine("--------------------------------");
    await client.JoinGameAsync(botConfig.RoomId);
    Console.WriteLine($"Joined room: {botConfig.RoomId}");

    // Keep the application running
    await host.RunAsync();
}
catch (Exception ex)
{
    Console.WriteLine($"Error: {ex.Message}");
}
