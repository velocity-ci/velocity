module Data.Event exposing (Event(..))

{- Events are basically lifecycle events for any particular "Model". Usually used in conjunction with a socket -}


type Event a
    = Created a
    | Completed a
