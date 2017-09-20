module Data.Helpers exposing (stringToDateTime)

import Json.Decode as Decode exposing (Decoder)
import Time.DateTime as DateTime exposing (DateTime)


-- SERIALIZATION --


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
