module Context exposing (Context, initContext)


type alias Context =
    { apiUrlBase : String
    , wsUrl : String
    }


initContext : String -> Context
initContext apiUrlBase =
    let
        wsUrlBase =
            if String.startsWith "http" apiUrlBase then
                "ws" ++ String.dropLeft 4 apiUrlBase
            else
                -- Not an http API URL - this will fail pretty quickly
                apiUrlBase
    in
        { apiUrlBase = apiUrlBase
        , wsUrl = wsUrlBase ++ "/ws"
        }
