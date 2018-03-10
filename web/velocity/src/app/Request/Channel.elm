module Request.Channel exposing (Error(..), errorDecoder, joinChannels, leaveChannels)

import Phoenix.Socket as Socket exposing (Socket)
import Phoenix.Channel as Channel exposing (Channel)
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Dict exposing (Dict)


type Error
    = AccessDenied


errorDecoder : Decoder Error
errorDecoder =
    Decode.at [ "message" ]
        (Decode.string
            |> Decode.andThen
                (\status ->
                    case status of
                        "access denied" ->
                            Decode.succeed AccessDenied

                        unknown ->
                            Decode.fail <| "Unknown error message: " ++ unknown
                )
        )


joinChannels :
    Socket toMsg
    -> (msg -> toMsg)
    -> Dict String (List ( String, Encode.Value -> msg ))
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
joinChannels socket toMsg channelsDict =
    channelsDict
        |> Dict.toList
        |> List.foldl (joinChannel toMsg) ( socket, Cmd.none )


joinChannel :
    (msg -> toMsg)
    -> ( String, List ( String, Encode.Value -> msg ) )
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
    -> ( Socket toMsg, Cmd (Socket.Msg toMsg) )
joinChannel toMsg ( channelName, events ) ( socket, cmd ) =
    let
        channel =
            channelName
                |> Channel.init
                |> Channel.map toMsg

        ( channelSocket, socketCmd ) =
            Socket.join channel socket

        foldEvents ( event, msg ) s =
            Socket.on event channel.name (msg >> toMsg) s
    in
        List.foldl foldEvents channelSocket events
            ! [ cmd, socketCmd ]


leaveChannels :
    (Socket.Msg msg -> toMsg)
    -> List String
    -> Socket msg
    -> ( Socket msg, Cmd toMsg )
leaveChannels toMsg channels socket =
    List.foldl
        (\channel ( socket, cmd ) ->
            let
                ( leaveSocket, leaveCmd ) =
                    Socket.leave channel socket
            in
                leaveSocket ! [ cmd, Cmd.map toMsg leaveCmd ]
        )
        ( socket, Cmd.none )
        channels
