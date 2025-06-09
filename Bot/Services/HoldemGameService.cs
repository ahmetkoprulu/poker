using System.Drawing;
using Bot.Models;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using Pastel;
using Websocket.Models;

namespace Bot.Services;

public interface IGameSerivice<TState, TAction>
{
    public int? GetAvailablePosition();
    public void SetState(JToken data);
    public void SetState(TState state);
    public void HandleGameActionMessage(MessageGameAction data);
    public void HandleGameActionMessage(TAction data);
}

public class HoldemGameService : IGameSerivice<HoldemGameState, HoldemActionMessage>
{
    public HoldemGameState State { get; set; }
    public RoomState RoomState { get; set; }

    public int? GetAvailablePosition()
    {
        for (var i = 0; i < RoomState.MaxGamePlayers; i++)
        {
            if (State.Players.FirstOrDefault(p => p.Position == i) == null)
                return i;
        }

        return null;
    }

    public void HandleGameActionMessage(MessageGameAction data)
    {
        var message = data.Data.ToObject<HoldemActionMessage>();
        if (message == null) throw new Exception("Invalid action message");

        HandleGameActionMessage(message);
    }

    public void HandleGameActionMessage(HoldemActionMessage data)
    {

    }

    public void SetState(HoldemGameState state)
    {
        State = state;
    }

    public void SetState(JToken data)
    {
        var state = data.ToObject<HoldemGameState>();
        if (state == null) return;

        SetState(state);
    }
}