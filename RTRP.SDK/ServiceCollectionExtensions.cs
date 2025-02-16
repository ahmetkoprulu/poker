using Microsoft.Extensions.DependencyInjection;

namespace RTRP.SDK;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddRTRPClient(this IServiceCollection services, string baseUrl)
    {
        services.AddHttpClient<IAuthClient, AuthClient>(client =>
        {
            client.BaseAddress = new Uri(baseUrl);
        });

        services.AddHttpClient<IEventClient, EventClient>(client =>
        {
            client.BaseAddress = new Uri(baseUrl);
        });

        services.AddHttpClient<IPlayerClient, PlayerClient>(client =>
        {
            client.BaseAddress = new Uri(baseUrl);
        });

        return services;
    }
}