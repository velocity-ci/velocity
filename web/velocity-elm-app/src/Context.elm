module Context exposing (Context, baseUrl, device, joinChannel, on, socket, socketSubscriptions, start, updateSocket, windowResize, wsUrl)

{-| The runtime context of the application.
-}

import Api exposing (BaseUrl, Cred)
import Dict exposing (Dict)
import Element exposing (Device)
import Email exposing (Email)
import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode exposing (Value)
import Phoenix.Channel as Channel exposing (Channel)
import Phoenix.Socket as Socket exposing (Socket)
import Project exposing (Project)
import Task exposing (Task)
import Username exposing (Username)



-- TYPES


type Context msg
    = Context BaseUrl Device (Socket msg)


start : BaseUrl -> { width : Int, height : Int } -> Context msg
start baseUrl_ dimensions =
    let
        device_ =
            Element.classifyDevice dimensions

        socket_ =
            Socket.init
                |> Socket.withDebug
    in
    Context baseUrl_ device_ socket_


type Update msg
    = SocketMsg (Socket.Msg msg)



-- INFO


baseUrl : Context msg -> BaseUrl
baseUrl (Context val _ _) =
    val


wsUrl : Context msg -> String
wsUrl (Context val _ _) =
    Api.toWsEndpoint val ++ "/v1/ws"


device : Context msg -> Device
device (Context _ val _) =
    val


socket : Context msg -> Socket msg
socket (Context _ _ val) =
    val



-- CHANGES


windowResize : { width : Int, height : Int } -> Context msg -> Context msg
windowResize dimensions (Context baseUrl_ _ socket_) =
    Context baseUrl_ (Element.classifyDevice dimensions) socket_



-- SOCKET


on : String -> String -> (Encode.Value -> msg) -> Context msg -> Context msg
on eventName channelName onReceive (Context baseUrl_ device_ socket_) =
    socket_
        |> Socket.on eventName channelName onReceive
        |> Context baseUrl_ device_


joinChannel : Channel msg -> Cred -> Context msg -> ( Context msg, Cmd (Socket.Msg msg) )
joinChannel channel cred (Context baseUrl_ device_ socket_) =
    let
        payload =
            Api.credPayload cred channel

        ( updatedSocket, socketCmd ) =
            Socket.join payload socket_
    in
    ( Context baseUrl_ device_ updatedSocket
    , socketCmd
    )


updateSocket : Socket.Msg msg -> Context msg -> ( Context msg, Cmd (Socket.Msg msg) )
updateSocket subMsg (Context baseUrl_ device_ socket_) =
    let
        ( updatedSocket, socketCmd ) =
            Socket.update subMsg socket_
    in
    ( Context baseUrl_ device_ updatedSocket
    , socketCmd
    )


socketSubscriptions : (Socket.Msg msg -> msg) -> Context msg -> Sub msg
socketSubscriptions toMsg (Context _ _ socket_) =
    Socket.listen socket_ toMsg
