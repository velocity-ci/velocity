module Request.Channel exposing (joinChannels, leaveChannels)

import Data.AuthToken as AuthToken exposing (AuthToken)
import Dict exposing (Dict)
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Phoenix.Channel as Channel exposing (Channel)
import Phoenix.Socket as Socket exposing (Socket)
import Request.Errors exposing (Error(..), HandledError(..))
import Util exposing ((=>))


joinChannels :
    Socket toMsg
    -> Maybe AuthToken
    -> (Request.Errors.Error Error -> toMsg)
    -> (msg -> toMsg)
    -> Dict String (List ( String, Encode.Value -> msg ))
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
joinChannels socket maybeAuthToken errorHandler toMsg channels =
    let
        join =
            joinChannel toMsg maybeAuthToken errorHandler
    in
        channels
            |> Dict.toList
            |> List.foldl join ( socket, Cmd.none )


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
                ( leaveSocket
                , Cmd.batch [ cmd, Cmd.map toMsg leaveCmd ]
                )
    in
        List.foldl leaveChannel ( socket, Cmd.none ) channels


withAuthToken :
    Channel msg
    -> AuthToken
    -> Channel msg
withAuthToken channel authToken =
    let
        payload =
            Encode.object [ "token" => AuthToken.encode authToken ]
    in
        Channel.withPayload payload channel


joinChannel :
    (msg -> toMsg)
    -> Maybe AuthToken
    -> (Request.Errors.Error Error -> toMsg)
    -> ( String, List ( String, Encode.Value -> msg ) )
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
joinChannel toMsg maybeAuthToken errorHandler ( channelName, events ) ( socket, cmd ) =
    let
        onJoinError value =
            value
                |> Decode.decodeValue errorDecoder
                |> handleChannelError
                |> errorHandler

        channel_ =
            channelName
                |> Channel.init
                |> Channel.map toMsg
                |> Channel.onJoinError onJoinError

        channel =
            maybeAuthToken
                |> Maybe.map (withAuthToken channel_)
                |> Maybe.withDefault channel_

        ( channelSocket, socketCmd ) =
            Socket.join channel socket

        foldEvent ( event, msg ) s =
            Socket.on event channel.name (msg >> toMsg) s
    in
        ( List.foldl foldEvent channelSocket events
        , Cmd.batch [ cmd, socketCmd ]
        )



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
