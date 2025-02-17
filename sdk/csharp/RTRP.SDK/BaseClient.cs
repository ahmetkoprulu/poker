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
        private string _baseUrl;
        private HttpClient _httpClient;

        public string BaseUrl
        {
            get { return _baseUrl; }
            set { _baseUrl = value; }
        }

        public HttpClient HttpClient
        {
            get { return _httpClient; }
            protected set { _httpClient = value; }
        }

        protected BaseClient()
        {
        }

        protected BaseClient(HttpClient httpClient, string baseUrl)
        {
            _httpClient = httpClient;
            _baseUrl = baseUrl;
        }
    }
}
