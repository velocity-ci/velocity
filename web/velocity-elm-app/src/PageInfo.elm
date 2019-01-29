module PageInfo exposing
    ( PageInfo
    , endCursor
    , hasNextPage
    , hasPreviousPage
    , selectionSet
    , startCursor
    , init
    )

import Api.Compiled.Object
import Api.Compiled.Object.PageInfo as PageInfo
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)



-- SERIALIZATION --


type PageInfo
    = PageInfo Internals


type alias Internals =
    { startCursor : Maybe String
    , endCursor : Maybe String
    , hasNextPage : Bool
    , hasPreviousPage : Bool
    }

init : PageInfo
init=
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
        |> SelectionSet.with PageInfo.startCursor
        |> SelectionSet.with PageInfo.endCursor
        |> SelectionSet.with PageInfo.hasNextPage
        |> SelectionSet.with PageInfo.hasPreviousPage



-- INFO


startCursor : PageInfo -> Maybe String
startCursor (PageInfo internals) =
    internals.startCursor


endCursor : PageInfo -> Maybe String
endCursor (PageInfo internals) =
    internals.endCursor


hasNextPage : PageInfo -> Bool
hasNextPage (PageInfo internals) =
    internals.hasNextPage


hasPreviousPage : PageInfo -> Bool
hasPreviousPage (PageInfo internals) =
    internals.hasPreviousPage
