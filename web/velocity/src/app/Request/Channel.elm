module Request.Channel exposing (joinChannels, leaveChannels)

import Phoenix.Socket as Socket exposing (Socket)
import Phoenix.Channel as Channel exposing (Channel)
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Dict exposing (Dict)
import Request.Errors exposing (HandledError(..), Error(..))
import Data.AuthToken as AuthToken exposing (AuthToken)
import Util exposing ((=>))


joinChannels :
    Socket toMsg
    -> (Request.Errors.Error Error -> toMsg)
    -> (msg -> toMsg)
    -> Dict String (List ( String, Encode.Value -> msg ))
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
joinChannels socket errorHandler toMsg channelsDict =
    channelsDict
        |> Dict.toList
        |> List.foldl (joinChannel toMsg errorHandler) ( socket, Cmd.none )


leaveChannels :
    (Socket.Msg msg -> toMsg)
    -> List String
    -> Socket msg
    -> ( Socket msg, Cmd toMsg )
leaveChannels toMsg channels socket =
    let
        leaveChannel channel ( socket, cmd ) =
            let
                ( leaveSocket, leaveCmd ) =
                    Socket.leave channel socket
            in
                leaveSocket ! [ cmd, Cmd.map toMsg leaveCmd ]
    in
        List.foldl leaveChannel ( socket, Cmd.none ) channels


withAuthToken :
    ( Channel msg, Cmd toMsg )
    -> AuthToken
    -> ( Channel msg, Cmd toMsg )
withAuthToken ( channel, cmd ) authToken =
    let
        payload =
            Encode.object [ "token" => AuthToken.encode authToken ]
    in
        Channel.withPayload payload channel => cmd


joinChannel :
    (msg -> toMsg)
    -> (Request.Errors.Error Error -> toMsg)
    -> ( String, List ( String, Encode.Value -> msg ) )
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
joinChannel toMsg errorHandler ( channelName, events ) ( socket, cmd ) =
    let
        onJoinError value =
            value
                |> Decode.decodeValue errorDecoder
                |> handleChannelError
                |> errorHandler

        channel =
            channelName
                |> Channel.init
                |> Channel.map toMsg
                |> Channel.onJoinError onJoinError

        ( channelSocket, socketCmd ) =
            Socket.join channel socket

        foldEvent ( event, msg ) s =
            Socket.on event channel.name (msg >> toMsg) s
    in
        List.foldl foldEvent channelSocket events
            ! [ cmd, socketCmd ]



-- ERRORS


type Error
    = AccessDenied


mapError : String -> Maybe Error
mapError errorMsg =
    case errorMsg of
        "access denied" ->
            Just AccessDenied

        _ ->
            Nothing


errorDecoder : Decoder Error
errorDecoder =
    Decode.at [ "message" ]
        (Decode.string
            |> Decode.andThen
                (\errorMsg ->
                    errorMsg
                        |> mapError
                        |> Maybe.map Decode.succeed
                        |> Maybe.withDefault (Decode.fail <| "Unknown error message: " ++ errorMsg)
                )
        )


handleChannelError : Result String Error -> Request.Errors.Error Error
handleChannelError res =
    case res of
        Ok AccessDenied ->
            HandledError Unauthorized

        Err _ ->
            HandledError Unauthorized
