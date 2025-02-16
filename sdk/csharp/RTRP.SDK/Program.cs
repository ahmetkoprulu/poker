namespace RTRP.SDK;

using RTRP.SDK;
using RTRP.SDK.Models;

// Create an HttpClient (preferably using dependency injection)
var httpClient = new HttpClient();

// Create the API clients
var authClient = new AuthClient(httpClient, "http://localhost:8000");
var eventClient = new EventClient(httpClient, "http://localhost:8000");
var playerClient = new PlayerClient(httpClient, "http://localhost:8000");

// Use the clients
try {
    // Login
    var loginResult = await authClient.LoginAsync(new LoginRequest
    {
        Provider = SocialNetwork.Email,
        Identifier = "user@example.com",
        Secret = "password"
    });

// Get events
var events = await eventClient.ListEventsAsync();

// Get player profile
var player = await playerClient.GetMyPlayerAsync();
}
catch (ApiException ex)
{
    Console.WriteLine($"API Error: {ex.Message}");
}