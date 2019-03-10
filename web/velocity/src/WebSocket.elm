port module WebSocket exposing (listen, open, send)

import Json.Decode as Decode
import Json.Encode as Encode


port open_ : Encode.Value -> Cmd msg


port opened_ : (Decode.Value -> msg) -> Sub msg


port onMessage : (String -> msg) -> Sub msg


port send_ : String -> Cmd msg



--port listenSub : (Decode.Value -> msg) -> Sub msg


open : String -> Cmd msg
open url =
    open_ (Encode.string url)


listen : (String -> msg) -> Sub msg
listen tagger =
    onMessage tagger


send : String -> Cmd msg
send message =
    send_ message
