module Data.PaginatedList exposing (PaginatedList, Paginated(..), decoder)

import Data.Project as Project exposing (Project)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (decode, required)


type alias PaginatedList a =
    Paginated (List a)


type Paginated a
    = Paginated { total : Int, results : a }



-- SERIALIZATION --


decoder : Decoder a -> Decoder (PaginatedList a)
decoder decoder =
    decode fromList
        |> required "total" Decode.int
        |> required "data" (Decode.list decoder)



-- HELPERS --


fromList : Int -> List a -> PaginatedList a
fromList a b =
    Paginated
        { total = a
        , results = b
        }
