port module Ports exposing (..)

import Json.Encode exposing (Value)


port storeSession : Maybe String -> Cmd msg


port onSessionChange : (Value -> msg) -> Sub msg


port parseGitUrl : String -> Cmd msg


port onGitUrlParsed : (Value -> msg) -> Sub msg


port onScrolledToBottom : (Value -> msg) -> Sub msg
