using Bot.Models;
using Bot.Extensions;
using Bot.Services;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Websocket.Services;

namespace Bot;

public class Program
{
    public static async Task Main(string[] args)
    {
        var host = CreateHostBuilder(args).Build();
        var botService = host.Services.GetRequiredService<BotService>();
        var webSocketClient = host.Services.GetRequiredService<IWebSocketClient>();
        var messageHandlerRegistry = host.Services.GetRequiredService<MessageHandlerRegistry>();
        var logger = host.Services.GetRequiredService<ILogger<Program>>();

        try
        {
            // messageHandlerRegistry.On<string>(MessageType.Error, async (error) =>
            // {
            //     logger.LogError("Received error: {Error}", error);
            // });

            // messageHandlerRegistry.OnError(async (error) =>
            // {
            //     logger.LogError("WebSocket error: {Error}", error);
            // });

            // _wsUrl = configuration["WebSocketUrl"] ?? "ws://localhost:8080/ws";

            // await webSocketClient.StartAsync();
            logger.LogInformation("Bot started successfully");

            // Join a game room
            var roomId = host.Services.GetRequiredService<IConfiguration>()["RoomId"];
            var position = host.Services.GetRequiredService<IConfiguration>()["Position"];
            if (!string.IsNullOrEmpty(roomId))
            {
                await webSocketClient.JoinGameAsync(roomId, int.Parse(position));
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
                services.AddSingleton<BotService>();
                services.AddSingleton<IWebSocketClient, WebSocketClient>();
                services.AddSingleton<MessageHandlerRegistry>();

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
