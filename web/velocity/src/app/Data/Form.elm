module Data.Form exposing (FormField)


type alias FormField a =
    { value : String
    , dirty : Bool
    , field : a
    }
