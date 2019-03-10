module Edge exposing
    ( Cursor
    , Edge
    , afterQueryParser
    , cursorFromString
    , cursorSelectionSet
    , cursorString
    , fromSelectionSet
    , node
    )

import Api.Compiled.Object
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)
import Url.Builder
import Url.Parser.Query as QueryParser



-- Info


node : Edge a -> a
node (Edge internals) =
    internals.node


fromSelectionSet : String -> a -> Edge a
fromSelectionSet cursor node_ =
    Edge { cursor = Cursor cursor, node = node_ }


type Edge a
    = Edge (Internals a)


type Cursor
    = Cursor String


type alias Internals a =
    { cursor : Cursor
    , node : a
    }


cursorSelectionSet : SelectionSet (Maybe String) typeLock -> SelectionSet (Maybe Cursor) typeLock
cursorSelectionSet =
    SelectionSet.map (Maybe.map Cursor)


afterQueryParser : QueryParser.Parser (Maybe String) -> QueryParser.Parser (Maybe Cursor)
afterQueryParser =
    QueryParser.map (Maybe.map Cursor)


cursorString : Cursor -> String
cursorString (Cursor val) =
    val


cursorFromString : String -> Cursor
cursorFromString =
    Cursor
