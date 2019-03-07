module Build.Status exposing (Status, decoder)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline
import Json.Encode as Encode


type Status
    = Waiting
    | Failed
    | Running
    | Success


decoder : Decoder Status
decoder =
    Decode.string
        |> Decode.andThen
            (\status ->
                case status of
                    "waiting" ->
                        Decode.succeed Waiting

                    "failed" ->
                        Decode.succeed Failed

                    "running" ->
                        Decode.succeed Running

                    "success" ->
                        Decode.succeed Success

                    unknown ->
                        Decode.fail <| "Unknown status: " ++ unknown
            )
