using System.Net.Http;
using System.Text;
using Newtonsoft.Json;

namespace RTRP.SDK;

public interface IBaseClient
{
    string BaseUrl { get; set; }
    HttpClient HttpClient { get; }
}

public class BaseClient : IBaseClient
{
    public string BaseUrl { get; set; }
    public HttpClient HttpClient { get; }

    public BaseClient(HttpClient httpClient, string baseUrl)
    {
        HttpClient = httpClient;
        BaseUrl = baseUrl;
    }

    protected async Task<T> SendAsync<T>(HttpRequestMessage request, CancellationToken cancellationToken = default)
    {
        var response = await HttpClient.SendAsync(request, cancellationToken);
        var content = await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
        {
            throw new ApiException(response.StatusCode, content);
        }

        return JsonConvert.DeserializeObject<T>(content)!;
    }
}

public class ApiException : Exception
{
    public System.Net.HttpStatusCode StatusCode { get; }

    public ApiException(System.Net.HttpStatusCode statusCode, string message) : base(message)
    {
        StatusCode = statusCode;
    }
}