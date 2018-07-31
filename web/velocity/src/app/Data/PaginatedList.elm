module Data.PaginatedList exposing (PaginatedList, Paginated, decoder, results, updateResults, total)

import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required, optional)


type alias PaginatedList a =
    Paginated (List a)


type Paginated a
    = Paginated { total : Int, results : a }



-- SERIALIZATION --


decoder : Decoder a -> Decoder (PaginatedList a)
decoder decoder =
    decode fromList
        |> required "total" Decode.int
        |> optional "data" (Decode.list decoder) []



-- HELPERS --


updateResults : PaginatedList a -> (List a -> List a) -> PaginatedList a
updateResults paginatedList updateFn =
    let
        newResults =
            updateFn (results paginatedList)
    in
        Paginated { total = List.length newResults, results = newResults }


results : PaginatedList a -> List a
results (Paginated { results }) =
    results


total : PaginatedList a -> Int
total (Paginated { total }) =
    total


fromList : Int -> List a -> PaginatedList a
fromList a b =
    Paginated
        { total = a
        , results = b
        }
