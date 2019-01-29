module Edge exposing (Edge, fromSelectionSet, node)


-- Info

node : Edge a -> a
node (Edge internals) =
    internals.node

fromSelectionSet : String -> a -> Edge a
fromSelectionSet cursor node_ =
    Edge { cursor = cursor, node = node_ }


type Edge a
    = Edge (Internals a)


type alias Internals a =
    { cursor : String
    , node : a
    }


