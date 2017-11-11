module Socket.Helpers exposing (..)

import Json.Decode as Decode exposing (field)
import Json.Encode as JE


type alias Message =
    { event : String
    , topic : String
    , payload : Decode.Value
    , ref : Maybe Int
    }


maybeInt : Maybe Int -> JE.Value
maybeInt maybe =
    case maybe of
        Just num ->
            JE.int num

        Nothing ->
            JE.null


nullOrInt : Decode.Decoder (Maybe Int)
nullOrInt =
    Decode.oneOf
        [ Decode.null Nothing
        , Decode.map Just Decode.int
        ]


messageDecoder : Decode.Decoder Message
messageDecoder =
    Decode.map4 Message
        (field "event" Decode.string)
        (field "topic" Decode.string)
        (field "payload" Decode.value)
        (field "ref" nullOrInt)


messageEncoder : Message -> JE.Value
messageEncoder { topic, event, payload, ref } =
    JE.object
        [ ( "type", JE.string "subscribe" )
        , ( "route", JE.string topic )
        ]


encodeMessage : Message -> String
encodeMessage =
    messageEncoder >> JE.encode 0
