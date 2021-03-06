module Project.Task.Slug exposing (Slug, decoder, urlParser)

import Api.Compiled.Object
import Api.Compiled.Object.Task as Task
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import Json.Decode as Decode exposing (Decoder)
import Url.Parser exposing (Parser)



-- TYPES


type Slug
    = Slug String



-- CREATE


urlParser : Parser (Slug -> a) a
urlParser =
    Url.Parser.custom "SLUG" (\str -> Just (Slug str))


decoder : Decoder Slug
decoder =
    Decode.map Slug Decode.string
