module Socket.Channel exposing (..)

import Json.Encode as JE


type alias Channel msg =
    { name : String
    , payload : JE.Value
    , state : State
    , onClose : Maybe (JE.Value -> msg)
    , onError : Maybe (JE.Value -> msg)
    , onJoin : Maybe (JE.Value -> msg)
    , onJoinError : Maybe (JE.Value -> msg)
    , joinRef : Int
    , leaveRef : Int
    }


init : String -> Channel msg
init name =
    { name = name
    , payload = (JE.object [])
    , state = Closed
    , onClose = Nothing
    , onError = Nothing
    , onJoin = Nothing
    , onJoinError = Nothing
    , joinRef = -1
    , leaveRef = -1
    }


{-| -}
map : (msg1 -> msg2) -> Channel msg1 -> Channel msg2
map fn channel =
    { channel
        | onClose = Maybe.map ((<<) fn) channel.onClose
        , onError = Maybe.map ((<<) fn) channel.onError
        , onJoin = Maybe.map ((<<) fn) channel.onJoin
        , onJoinError = Maybe.map ((<<) fn) channel.onJoinError
    }


type State
    = Closed
    | Errored
    | Joined
    | Joining
    | Leaving
