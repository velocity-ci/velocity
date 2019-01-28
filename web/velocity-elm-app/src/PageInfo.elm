module PageInfo exposing
    ( PageInfo
    , endCursor
    , hasNextPage
    , hasPreviousPage
    , selectionSet
    , startCursor
    )

import Api.Compiled.Object
import Api.Compiled.Object.PageInfo as PageInfo
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, with)



-- SERIALIZATION --


type PageInfo
    = PageInfo Internals


type alias Internals =
    { startCursor : String
    , endCursor : String
    , hasNextPage : Bool
    , hasPreviousPage : Bool
    }


selectionSet : SelectionSet PageInfo Api.Compiled.Object.PageInfo
selectionSet =
    SelectionSet.succeed PageInfo
        |> with internalsSelectionSet


internalsSelectionSet : SelectionSet Internals Api.Compiled.Object.PageInfo
internalsSelectionSet =
    SelectionSet.succeed Internals
        |> SelectionSet.with (SelectionSet.nonNullOrFail PageInfo.startCursor)
        |> SelectionSet.with (SelectionSet.nonNullOrFail PageInfo.endCursor)
        |> SelectionSet.with PageInfo.hasNextPage
        |> SelectionSet.with PageInfo.hasPreviousPage



-- INFO


startCursor : PageInfo -> String
startCursor (PageInfo internals) =
    internals.startCursor


endCursor : PageInfo -> String
endCursor (PageInfo internals) =
    internals.endCursor


hasNextPage : PageInfo -> Bool
hasNextPage (PageInfo internals) =
    internals.hasNextPage


hasPreviousPage : PageInfo -> Bool
hasPreviousPage (PageInfo internals) =
    internals.hasPreviousPage
