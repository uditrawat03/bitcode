package sidebar

// import "github.com/gdamore/tcell/v2"

// // Handle key events
// // Handle key events
// func (sb *Sidebar) HandleKey(ev *tcell.EventKey) {
// 	switch ev.Key() {
// 	case tcell.KeyUp:
// 		sb.Scroll(-1)
// 	case tcell.KeyDown:
// 		sb.Scroll(1)
// 	case tcell.KeyEnter, tcell.KeyRight:
// 		node := sb.Tree.Nodes[sb.Selected]
// 		if node.IsDir && !node.Expanded {
// 			sb.Tree.Toggle(node)
// 		} else if sb.onFileOpen != nil {
// 			sb.onFileOpen(node.Path)
// 		}
// 	case tcell.KeyLeft:
// 		node := sb.Tree.Nodes[sb.Selected]
// 		if node.IsDir && node.Expanded {
// 			sb.Tree.Toggle(node)
// 		} else if node.Parent != nil {
// 			// Select parent
// 			for i, n := range sb.Tree.Nodes {
// 				if n == node.Parent {
// 					sb.Selected = i
// 					sb.Scroll(0)
// 					break
// 				}
// 			}
// 		}
// 	}
// }
