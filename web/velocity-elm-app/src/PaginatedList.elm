module PaginatedList exposing (PaginatedList, decoder, fromList, map, params, total, values)

import Html exposing (Html, a, li, text, ul)
import Html.Attributes exposing (class, classList, href)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (optional, required)
import Url.Builder exposing (QueryParameter)


-- TYPES


type PaginatedList a
    = PaginatedList
        { values : List a
        , total : Int
        }



-- INFO


values : PaginatedList a -> List a
values (PaginatedList info) =
    info.values


total : PaginatedList a -> Int
total (PaginatedList info) =
    info.total



-- CREATE


fromList : Int -> List a -> PaginatedList a
fromList totalCount list =
    PaginatedList { values = list, total = totalCount }



-- TRANSFORM


map : (a -> a) -> PaginatedList a -> PaginatedList a
map transform (PaginatedList info) =
    PaginatedList { info | values = List.map transform info.values }



-- PARAMS


{-| I decided to accept a record here so I don't mess up the argument order of the two Ints.
-}
params :
    { page : Int, resultsPerPage : Int }
    -> List QueryParameter
params { page, resultsPerPage } =
    let
        offset =
            (page - 1) * resultsPerPage
    in
        [ Url.Builder.string "limit" (String.fromInt resultsPerPage)
        , Url.Builder.string "offset" (String.fromInt offset)
        ]



-- SERIALIZATION --


decoder : Decoder a -> Decoder (PaginatedList a)
decoder decoder_ =
    Decode.succeed fromList
        |> required "total" Decode.int
        |> optional "data" (Decode.list decoder_) []
