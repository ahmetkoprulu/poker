# Product Requirements Document (PRD)
## Real-Time Poker Platform

### 1. Product Overview
#### 1.1 Purpose
A real-time multiplayer poker platform that allows users to play Texas Hold'em poker in a browser-based environment. The platform will support both active players and spectators, with a focus on real-time gameplay and user experience.

#### 1.2 Target Audience
- Casual poker players
- Online gaming enthusiasts
- People looking to learn poker
- Social players who want to play with friends

### 2. System Architecture
#### 2.1 Components
- **Frontend**: Next.js web application
- **Authentication Server**: Go/Gin REST API
- **Game Server**: Go WebSocket server
- **In-Memory State Management**: No persistent storage for game state

#### 2.2 Communication Flow
1. User authenticates via REST API
2. Frontend establishes WebSocket connection with game token
3. Direct real-time communication between frontend and game server
4. No database persistence for game states

### 3. Core Features

#### 3.1 Authentication
- **User Registration** (MVP)
  - Username
  - Password
  - Email (optional)
- **User Login** (MVP)
  - JWT-based authentication
  - Token validation for WebSocket connection

#### 3.2 Lobby System
- **Table Management** (MVP)
  - Single table with 5 seats
  - Spectator support
  - Real-time table status updates
- **Player Status** (MVP)
  - Seated/Standing
  - Active/Inactive
  - Current chip count

#### 3.3 Game Mechanics
- **Basic Gameplay** (MVP)
  - Texas Hold'em rules
  - 5 seats per table
  - Fixed starting chips (1000)
  - Fixed blind levels (10/20)
  
- **Betting Actions** (MVP)
  - Fold
  - Check
  - Call
  - Raise
  - All-in

- **Game Flow** (MVP)
  - Pre-flop
  - Flop
  - Turn
  - River
  - Showdown

#### 3.4 Real-Time Features
- **Game State Updates** (MVP)
  - Player actions
  - Card dealing
  - Pot updates
  - Timer updates
  - Current player indication

- **Spectator Mode** (MVP)
  - View ongoing games
  - Real-time game state updates
  - Join/Leave as spectator

### 4. Technical Requirements

#### 4.1 Frontend Requirements
- **Browser Support**
  - Modern browsers (Chrome, Firefox, Safari, Edge)
  - Mobile responsive design
  - WebSocket support

- **User Interface**
  - Poker table visualization
  - Player positions
  - Card display
  - Betting controls
  - Chip counts
  - Timer display

#### 4.2 Backend Requirements
- **WebSocket Server**
  - Concurrent connection handling
  - Real-time message broadcasting
  - Connection state management
  - Authentication validation

- **Game Logic**
  - Hand evaluation
  - Betting validation
  - Turn management
  - Pot calculation
  - Winner determination

#### 4.3 Performance Requirements
- **Latency**
  - Maximum 100ms action response time
  - Real-time updates within 50ms

- **Scalability**
  - Support for concurrent table
  - Minimum 100 concurrent connections

### 5. Message Types

#### 5.1 Client to Server
```json
{
    "join_table": {},
    "leave_table": {},
    "take_seat": {
        "seat": number
    },
    "leave_seat": {},
    "player_action": {
        "type": "fold|check|call|raise",
        "amount": number
    }
}
```

#### 5.2 Server to Client
```json
{
    "game_state": {
        "table_id": string,
        "players": [{
            "id": string,
            "username": string,
            "seat": number,
            "chips": number,
            "bet": number,
            "status": string
        }],
        "spectators": [{
            "id": string,
            "username": string
        }],
        "community_cards": string[],
        "pot": number,
        "current_bet": number,
        "current_player": string,
        "dealer_position": number,
        "status": string
    },
    "player_turn": {
        "player_id": string,
        "timeout": number
    },
    "hand_result": {
        "winners": [{
            "player_id": string,
            "amount": number,
            "hand": string
        }],
        "community_cards": string[],
        "player_cards": {
            "player_id": string,
            "cards": string[]
        }[]
    },
    "error": {
        "code": string,
        "message": string
    }
}
```

### 6. Game Rules

#### 6.1 Table Rules
- 5 seats maximum
- Fixed buy-in: 1000 chips
- Small blind: 10 chips
- Big blind: 20 chips
- No time bank
- 30-second action timer

#### 6.2 Betting Rules
- Minimum raise: Previous raise amount
- Maximum raise: Player's remaining chips
- All-in: Allowed
- No re-buy
- No add-on

### 7. Future Enhancements (Not in MVP)
1. Multiple tables
2. Chat system
3. Player statistics
4. Tournament support
5. Custom table settings
6. Hand history
7. Player rankings
8. Friend system
9. Achievement system
10. Different poker variants

### 8. Development Phases

#### Phase 1 - MVP (Current Focus)
- Basic authentication
- Single table support
- Core poker gameplay
- Spectator mode
- Real-time updates

#### Phase 2 (Future)
- Multiple tables
- Chat system
- Basic statistics
- Hand history

#### Phase 3 (Future)
- Tournament support
- Friend system
- Achievement system
- Advanced statistics 