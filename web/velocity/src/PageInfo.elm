module PageInfo exposing
    ( PageInfo
    , endCursor
    , hasNextPage
    , hasPreviousPage
    , init
    , selectionSet
    , startCursor
    )

import Api.Compiled.Object
import Api.Compiled.Object.PageInfo as PageInfo
import Edge
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)



-- SERIALIZATION --


type PageInfo
    = PageInfo Internals


type alias Internals =
    { startCursor : Maybe Edge.Cursor
    , endCursor : Maybe Edge.Cursor
    , hasNextPage : Bool
    , hasPreviousPage : Bool
    }


init : PageInfo
init =
    PageInfo
        { startCursor = Nothing
        , endCursor = Nothing
        , hasNextPage = False
        , hasPreviousPage = False
        }


selectionSet : SelectionSet PageInfo Api.Compiled.Object.PageInfo
selectionSet =
    SelectionSet.succeed PageInfo
        |> with internalsSelectionSet


internalsSelectionSet : SelectionSet Internals Api.Compiled.Object.PageInfo
internalsSelectionSet =
    SelectionSet.succeed Internals
        |> SelectionSet.with (Edge.cursorSelectionSet PageInfo.startCursor)
        |> SelectionSet.with (Edge.cursorSelectionSet PageInfo.endCursor)
        |> SelectionSet.with PageInfo.hasNextPage
        |> SelectionSet.with PageInfo.hasPreviousPage



-- INFO


startCursor : PageInfo -> Maybe Edge.Cursor
startCursor (PageInfo internals) =
    internals.startCursor


endCursor : PageInfo -> Maybe Edge.Cursor
endCursor (PageInfo internals) =
    internals.endCursor


hasNextPage : PageInfo -> Bool
hasNextPage (PageInfo internals) =
    internals.hasNextPage


hasPreviousPage : PageInfo -> Bool
hasPreviousPage (PageInfo internals) =
    internals.hasPreviousPage
