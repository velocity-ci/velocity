module Edge exposing (Edge, fromSelectionSet)


fromSelectionSet : String -> a -> Edge a
fromSelectionSet cursor node =
    Edge { cursor = cursor, node = node }


type Edge a
    = Edge (Internals a)


type alias Internals a =
    { cursor : String
    , node : a
    }
