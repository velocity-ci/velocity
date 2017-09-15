module Data.Project exposing (Project, decoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, decode, hardcoded, required)
import Time.DateTime as DateTime exposing (DateTime)


type alias Project =
    { id : String
    , key : String
    , name : String
    , repository : String
    , createdAt : DateTime
    , updatedAt : DateTime
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    decode Project
        |> required "id" Decode.string
        |> required "key" Decode.string
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" stringToDateTime
        |> required "updatedAt" stringToDateTime


stringToDateTime : Decoder DateTime
stringToDateTime =
    Decode.string
        |> Decode.andThen
            (\val ->
                case DateTime.fromISO8601 val of
                    Err err ->
                        Decode.fail err

                    Ok dateTime ->
                        Decode.succeed dateTime
            )
