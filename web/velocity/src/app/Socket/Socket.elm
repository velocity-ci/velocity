module Socket.Socket exposing (..)

import WebSocket
import Json.Decode as Decode exposing (field)
import Socket.Helpers exposing (Message, messageDecoder, encodeMessage)
import Socket.Channel as Channel exposing (Channel)
import Socket.Push as Push exposing (Push)
import Dict exposing (Dict)
import Json.Encode as JE


type alias Socket msg =
    { path : String
    , channels : Dict String (Channel msg)
    , pushes : Dict Int (Push msg)
    , ref : Int
    }


type Msg msg
    = NoOp
    | ExternalMsg msg


init : String -> Socket msg
init path =
    { path = path
    , channels = Dict.fromList []
    , pushes = Dict.fromList []
    , ref = 0
    }


{-| -}
update : Msg msg -> Socket msg -> ( Socket msg, Cmd (Msg msg) )
update msg socket =
    case msg of
        --        ChannelErrored channelName ->
        --            let
        --                channels =
        --                    Dict.update channelName (Maybe.map (setState Channel.Errored)) socket.channels
        --
        --                socket_ =
        --                    { socket | channels = channels }
        --            in
        --                ( socket_, Cmd.none )
        --
        --        ChannelClosed channelName ->
        --            case Dict.get channelName socket.channels of
        --                Just channel ->
        --                    let
        --                        channels =
        --                            Dict.insert channelName (setState Channel.Closed channel) socket.channels
        --
        --                        pushes =
        --                            Dict.remove channel.joinRef socket.pushes
        --
        --                        socket_ =
        --                            { socket | channels = channels, pushes = pushes }
        --                    in
        --                        ( socket_, Cmd.none )
        --
        --                Nothing ->
        --                    ( socket, Cmd.none )
        --
        --        ChannelJoined channelName ->
        --            case Dict.get channelName socket.channels of
        --                Just channel ->
        --                    let
        --                        channels =
        --                            Dict.insert channelName (setState Channel.Joined channel) socket.channels
        --
        --                        pushes =
        --                            Dict.remove channel.joinRef socket.pushes
        --
        --                        socket_ =
        --                            { socket | channels = channels, pushes = pushes }
        --                    in
        --                        ( socket_, Cmd.none )
        --
        --                Nothing ->
        --                    ( socket, Cmd.none )
        --
        --        Heartbeat _ ->
        --            heartbeat socket
        _ ->
            ( socket, Cmd.none )


listen : Socket msg -> (Msg msg -> msg) -> Sub msg
listen socket fn =
    (Sub.batch >> Sub.map (mapAll fn))
        [ internalMsgs socket
        , externalMsgs socket
        ]


mapAll : (Msg msg -> msg) -> Msg msg -> msg
mapAll fn internalMsg =
    case internalMsg of
        ExternalMsg msg ->
            msg

        _ ->
            fn internalMsg


join : Channel msg -> Socket msg -> ( Socket msg, Cmd (Msg msg) )
join channel socket =
    case Dict.get channel.name socket.channels of
        Just { state } ->
            if state == Channel.Joined || state == Channel.Joining then
                ( socket, Cmd.none )
            else
                joinChannel channel socket

        Nothing ->
            joinChannel channel socket


joinChannel : Channel msg -> Socket msg -> ( Socket msg, Cmd (Msg msg) )
joinChannel channel socket =
    let
        push_ =
            Push "phx_join" channel.name channel.payload channel.onJoin channel.onJoinError

        channel_ =
            { channel | state = Channel.Joining, joinRef = socket.ref }

        socket_ =
            { socket
                | channels = Dict.insert channel.name channel_ socket.channels
            }
    in
        push push_ socket_


{-| Pushes a message
    push_ = Phoenix.Push.init "new:msg" "rooms:lobby"
    (socket_, cmd) = push push_ socket
-}
push : Push msg -> Socket msg -> ( Socket msg, Cmd (Msg msg) )
push push_ socket =
    ( { socket
        | pushes = Dict.insert socket.ref push_ socket.pushes
        , ref = socket.ref + 1
      }
    , send socket push_.event push_.channel push_.payload
    )


send : Socket msg -> String -> String -> JE.Value -> Cmd (Msg msg)
send { path, ref } event channel payload =
    sendMessage path (Message event channel payload (Just ref))


sendMessage : String -> Message -> Cmd (Msg msg)
sendMessage path message =
    WebSocket.send path (encodeMessage message)


internalMsgs : Socket msg -> Sub (Msg msg)
internalMsgs socket =
    Sub.map (mapInternalMsgs socket) (phoenixMessages socket)


mapInternalMsgs : Socket msg -> Maybe Message -> Msg msg
mapInternalMsgs socket maybeMessage =
    NoOp


externalMsgs : Socket msg -> Sub (Msg msg)
externalMsgs socket =
    Sub.map (mapExternalMsgs socket) (phoenixMessages socket)


mapExternalMsgs : Socket msg -> Maybe Message -> Msg msg
mapExternalMsgs socket maybeMessage =
    NoOp


phoenixMessages : Socket msg -> Sub (Maybe Message)
phoenixMessages socket =
    WebSocket.listen socket.path decodeMessage


decodeMessage : String -> Maybe Message
decodeMessage =
    Decode.decodeString messageDecoder >> Result.toMaybe
