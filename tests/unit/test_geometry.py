from src.domain.geometry import Contour


def test_area_rectangle():
    c = Contour([(0, 0), (3200, 0), (3200, 4000), (0, 4000)])
    assert round(c.area_m2(), 2) == 12.8
