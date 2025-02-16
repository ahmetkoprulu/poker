using System.Net.Http;

namespace RTRP.SDK
{
    public interface IBaseClient
    {
        string BaseUrl { get; set; }
        HttpClient HttpClient { get; }
    }

    public abstract class BaseClient : IBaseClient
    {
        public string BaseUrl { get; set; }
        public HttpClient HttpClient { get; }

        protected BaseClient(HttpClient httpClient, string baseUrl)
        {
            HttpClient = httpClient;
            BaseUrl = baseUrl;
        }
    }
}