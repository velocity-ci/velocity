module Context exposing (Context, initContext)

import Navigation exposing (Location)


type alias Context =
    { apiUrlBase : String
    , wsUrl : String
    }


initContext : Location -> Context
initContext { protocol, hostname } =
    -- Reminder: location.host includes port if there is one;
    -- location.hostname does not.
    let
        isSecureProtocol =
            protocol == "https:"

        -- API host is currently web hostname + default port
        apiHost =
            hostname

        wsProtocol =
            if isSecureProtocol then
                "wss:"
            else
                "ws:"
    in
        { apiUrlBase =
            protocol ++ "//" ++ apiHost ++ "/v1"
        , wsUrl =
            wsProtocol ++ "//" ++ apiHost ++ "/v1/ws"
        }
