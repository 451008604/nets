package common

import pb "github.com/451008604/socketServerFrame/proto/bin"

// 坐标转换
func convert2Index(x, y int) int {

	// 判断坐标是否合法
	if x < 0 || x >= ItemSpaceWidth || y < 0 || y >= ItemSpaceHeight {
		return -1
	}

	// 返回结果
	return y*ItemSpaceWidth + x
}

// 判断坐标是否为空
func isEmpty(ItemList []*pb.PBItemData, x, y int) bool {

	// 判断坐标是否合法
	if x < 0 || x >= ItemSpaceWidth || y < 0 || y >= ItemSpaceHeight {
		return false
	}

	// 计算坐标
	idx := convert2Index(x, y)

	// 返回结果
	return ItemList[idx].GetItemID() < 1
}

// 寻找空格子(如果是螺旋式查找, 不包含src本身)
func GetItemSpace(ItemList []*pb.PBItemData, l2r bool, src int) int32 {
	if l2r {
		// 从左往右, 从上到下
		for j := 0; j < ItemSpaceHeight; j++ {
			for i := 0; i < ItemSpaceWidth; i++ {

				// 找到空格子, 直接返回index
				if isEmpty(ItemList, i, j) {
					return int32(convert2Index(i, j))
				}
			}
		}

	} else {
		// 计算src对应的<x, y>
		srcX := src % ItemSpaceWidth
		srcY := src / ItemSpaceWidth

		// 顺时针查找(正方形区域)
		for n := 1; n < ItemSpaceWidth || n < ItemSpaceHeight; n++ {

			// 顺序遍历src右上方
			for x, y := srcX, srcY-n; x <= srcX+n; x++ {
				// 找到空格子, 直接返回index
				if isEmpty(ItemList, x, y) {
					return int32(convert2Index(x, y))
				}
			}

			// 顺序遍历src右侧
			for x, y := srcX+n, srcY-n; y <= srcY+n; y++ {

				// 找到空格子, 直接返回index
				if isEmpty(ItemList, x, y) {
					return int32(convert2Index(x, y))
				}
			}

			// 倒序遍历src下方
			for x, y := srcX+n, srcY+n; x >= srcX-n; x-- {

				// 找到空格子, 直接返回index
				if isEmpty(ItemList, x, y) {
					return int32(convert2Index(x, y))
				}
			}

			// 倒序遍历src左侧
			for x, y := srcX-n, srcY+n; y >= srcY-n; y-- {
				// 找到空格子, 直接返回index
				if isEmpty(ItemList, x, y) {
					return int32(convert2Index(x, y))
				}
			}

			// 遍历src左上方
			for x, y := srcX-n, srcY-n; x <= srcX; x++ {
				// 找到空格子, 直接返回index
				if isEmpty(ItemList, x, y) {
					return int32(convert2Index(x, y))
				}
			}
		}
	}

	return -1
}
