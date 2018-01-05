module Views.Build exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStreamOutput)
import Ansi.Log


viewBuildContainer : Build -> List BuildStep -> Ansi.Log.Model -> Html msg
viewBuildContainer build steps output =
    div []
        [ h3 [] [ text ("build container " ++ Build.idToString build.id) ]
        , Ansi.Log.view output
        ]



--
--
--viewBuildOutput : BuildStreamOutput -> List (Html msg)
--viewBuildOutput { output } =
--    [ br [] []
--    , text output
--    ]
--
--
--viewBuildStep : BuildStep -> Html msg
--viewBuildStep step =
--    div []
--        [ text (BuildStep.idToString step.id) ]
