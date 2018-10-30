port module Ports exposing (onGitUrlParsed, onScrolledToBottom, onSessionChange, parseGitUrl, scrollIntoView, scrollTo, storeSession)

import Dom exposing (Id)
import Json.Encode exposing (Value)


port storeSession : Maybe String -> Cmd msg


port onSessionChange : (Value -> msg) -> Sub msg


port parseGitUrl : String -> Cmd msg


port onGitUrlParsed : (Value -> msg) -> Sub msg


port onScrolledToBottom : (Value -> msg) -> Sub msg


port scrollIntoView : Id -> Cmd msg


port scrollTo : ( Int, Int ) -> Cmd msg
