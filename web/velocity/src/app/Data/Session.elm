module Data.Session exposing (Session, attempt)

import Data.User as User exposing (User)
import Data.AuthToken exposing (AuthToken)
import Socket.Socket as Socket exposing (Socket)
import Util exposing ((=>))


type alias Session msg =
    { user : Maybe User
    , socket : Socket msg
    }



-- HELPERS --


attempt : String -> (AuthToken -> Cmd a) -> Session b -> ( List String, Cmd a )
attempt attemptedAction toCmd session =
    case Maybe.map .token session.user of
        Nothing ->
            [ "You have been signed out. Please sign back in to " ++ attemptedAction ++ "." ] => Cmd.none

        Just token ->
            [] => toCmd token
