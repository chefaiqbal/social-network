PRAGMA foreign_keys = ON;

CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    creator_id INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE group_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER REFERENCES groups(id),
    user_id INTEGER REFERENCES users(id),
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);



CREATE TABLE group_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER REFERENCES groups(id),              
    creator_id INTEGER REFERENCES users(id),                  
    title TEXT NOT NULL,                     
    description TEXT NOT NULL,                    
    event_date DATETIME NOT NULL,                
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);


CREATE TABLE group_event_RSVP (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER REFERENCES group_events(id),                 
    user_id INTEGER REFERENCES users(id),               
    rsvp_status TEXT NOT NULL CHECK (rsvp_status IN ('going', 'not going')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
);
