# SES Algorithm - Schiper-Eggli-Sandoz Protocol Implementation

Đây là một **bài thực hành/mô phỏng thuật toán trong hệ thống phân tán (Distributed Systems)**, cụ thể là triển khai thuật toán **SES (Schiper-Eggli-Sandoz)** bằng ngôn ngữ Go.

Sự khác biệt:
- **Distributed Systems (Hệ phân tán)**: Lĩnh vực nghiên cứu về cách nhiều tiến trình/máy tính phối hợp với nhau, giải quyết các vấn đề như đồng bộ thời gian, thứ tự tin nhắn, đồng thuận, v.v.

Project này thuộc về **Distributed Systems** - nó mô phỏng nhiều tiến trình (process) giao tiếp qua mạng TCP và sử dụng thuật toán SES để đảm bảo **causal message ordering** (thứ tự nhân quả của tin nhắn).

---

## Thuật toán SES là gì?

Thuật toán **Schiper-Eggli-Sandoz (1989)** giải quyết bài toán **Causal Message Ordering** trong hệ phân tán, nghĩa là:

> Nếu tin nhắn A được gửi **trước** tin nhắn B (và B phụ thuộc vào A), thì bên nhận phải **deliver (xử lý) A trước B**, bất kể thứ tự chúng đến qua mạng.

Trong mạng thực tế, tin nhắn có thể đến **không đúng thứ tự** (out-of-order) do độ trễ mạng khác nhau. SES giải quyết điều này bằng cách sử dụng **Vector Clock** (đồng hồ vector) để theo dõi quan hệ nhân quả giữa các sự kiện.

---

## Chi tiết những gì đã làm

### 1. Cấu trúc dự án

```
SES-algorithm/
├── main.go                    # Điểm khởi chạy chương trình
├── go.mod                     # Go module definition
├── run.ps1                    # Script PowerShell để chạy nhiều process
├── SES/
│   ├── ses.go                 # Logic chính của thuật toán SES
│   ├── vector.clock.go        # Cài đặt Vector Clock
│   └── logical.clock.go       # Cài đặt Logical Clock (đồng hồ logic)
├── network/
│   ├── network.go             # Quản lý kết nối mạng TCP
│   ├── sender.go              # Worker gửi tin nhắn
│   ├── receiver.go            # Worker nhận tin nhắn
│   └── signal.handler.go      # Xử lý tín hiệu thoát chương trình
├── constant/
│   └── constant.go            # Các hằng số cấu hình
└── static/logs/               # Log ghi lại hoạt động gửi/nhận
    ├── 02_process/            # Log khi chạy 2 process
    │   ├── 00__sender.log
    │   ├── 00__receiver.log
    │   ├── 01__sender.log
    │   └── 01__receiver.log
    └── 07_process/            # Log khi chạy 7 process
        ├── 00__sender.log ~ 06__sender.log
        └── 00__receiver.log ~ 06__receiver.log
```

### 2. Logical Clock (`SES/logical.clock.go`)

Đây là đơn vị cơ bản nhất - một **đồng hồ logic** cho mỗi process:

- Mỗi `LogicClock` là một mảng `Clock[]int` có kích thước = số lượng process.
- **Khởi tạo**: Process của chính nó thì `Clock = [0, 0, ..., 0]`, còn các process khác thì `Clock = [-1, -1, ..., -1]` (giá trị `-1` đại diện cho "chưa biết thông tin gì").
- **`Increase()`**: Tăng giá trị clock tại vị trí của process hiện tại lên 1 (đại diện cho một sự kiện mới xảy ra).
- **`UpdateClock(other)`**: Merge (hợp nhất) clock với clock từ process khác bằng cách lấy `max()` từng phần tử.
- **`IsNull()`**: Kiểm tra xem clock có chứa giá trị `-1` không (chưa được khởi tạo).
- **`Serialize()`/`Deserialize()`**: Chuyển đổi clock thành byte array để gửi qua mạng và ngược lại.

### 3. Vector Clock (`SES/vector.clock.go`)

**Vector Clock** là một ma trận các Logical Clock:

- Mỗi process `P_i` duy trì một `VectorClock` gồm N vectors (N = số process).
- `Vectors[i]` = kiến thức của process hiện tại về trạng thái clock của process `i`.
- `Vectors[InstanceID]` = clock riêng của chính process đó (luôn cập nhật nhất).

Các thao tác:
- **`Increase()`**: Tăng clock riêng của process hiện tại.
- **`Merge(sourceVC, sourceID, destinationID)`**: Cập nhật kiến thức từ vector clock của process nguồn.
- **`SelfMerge(sourceID, destinationID)`**: Sao chép clock riêng vào vị trí của process đích (dùng khi gửi tin nhắn để ghi nhận "tôi đã cho process đích biết trạng thái của tôi đến đây").
- **`SerializeVectorClock()`/`DeserializeVectorClock()`**: Đóng gói toàn bộ vector clock + dữ liệu thành byte array để truyền qua mạng.

### 4. SES Core (`SES/ses.go`)

Đây là phần quan trọng nhất - logic gửi và nhận tin nhắn với đảm bảo thứ tự nhân quả:

#### Khi GỬI tin nhắn (`Send()`):
1. Khóa mutex (đồng bộ hóa).
2. **Tăng** clock riêng (`Increase`).
3. Ghi log gửi (sender log).
4. **Serialize** toàn bộ vector clock + nội dung tin nhắn thành byte array.
5. **SelfMerge**: Sao chép clock riêng vào vị trí của process đích, đánh dấu "process đích đã biết trạng thái của tôi tính đến thời điểm này".

#### Khi NHẬN tin nhắn (`Deliver()`):
1. Khóa mutex.
2. **Deserialize** để lấy lại vector clock của người gửi và nội dung tin nhắn.
3. Lấy `t_m` = kiến thức của người gửi về trạng thái của người nhận tại thời điểm gửi.
4. Lấy `timeProcess` = clock hiện tại của người nhận.
5. **Kiểm tra điều kiện deliver**: `t_m <= timeProcess` (tất cả phần tử của `t_m.Clock[]` đều `<=` phần tử tương ứng của `timeProcess.Clock[]`).

   - **Nếu thỏa mãn (DELIVER)**: Tin nhắn được deliver ngay lập tức, sau đó merge vector clock.
   - **Nếu KHÔNG thỏa (BUFFER)**: Tin nhắn được đưa vào hàng đợi (queue), vì có tin nhắn trước đó chưa đến.

6. Sau khi buffer, **lặp lại kiểm tra** tất cả tin nhắn trong queue - nếu có tin nhắn nào đã đủ điều kiện deliver thì deliver và xóa khỏi queue.

#### Merge khi deliver (`MergeSES()`):
- Cập nhật kiến thức về tất cả process khác (trừ chính mình và người gửi) bằng cách lấy `max()`.
- Cập nhật kiến thức mà người gửi có về chính mình vào clock riêng.
- Tăng clock riêng lên 1.

### 5. Network Layer (`network/`)

#### `network.go` - Quản lý mạng:
- Mỗi process lắng nghe trên cổng `PORT_OFFSET + instanceID` (ví dụ: process 0 nghe cổng 1200, process 1 nghe cổng 1201).
- Tất cả chạy trên `127.0.0.1` (localhost).
- `StartListening()`: Mở TCP server, chấp nhận kết nối đến, tạo `ReceiverWorker` cho mỗi kết nối.
- `StartSending()`: Tạo `SenderWorker` kết nối TCP đến tất cả process khác.

#### `sender.go` - Gửi tin nhắn:
- Mỗi `SenderWorker` kết nối TCP đến một process khác.
- Gửi **150 tin nhắn** (`MAX_MESSAGE = 150`) đến process đích.
- Mỗi tin nhắn có format: `"Message number X from process Y"`.
- **Delay ngẫu nhiên** 100ms - 1000ms giữa các lần gửi (mô phỏng mạng thực tế).
- Tin nhắn được đóng gói: `[4 byte kích thước][vector clock + nội dung]`.

#### `receiver.go` - Nhận tin nhắn:
- Đọc 4 byte đầu tiên để biết kích thước dữ liệu, sau đó đọc đúng số byte đó.
- **Cơ chế mô phỏng out-of-order**: Tin nhắn nhận được không xử lý ngay mà đưa vào mảng `Noise[]`. Sau đó, với xác suất 90% (`rand.Float32() > 0.1`), toàn bộ `Noise[]` được deliver **theo thứ tự ngược** (`i = len-1` đến `0`), mô phỏng tin nhắn đến không đúng thứ tự.
- Khi nhận đủ tất cả tin nhắn (`MAX_MESSAGE * (N-1)`), deliver toàn bộ `Noise[]` còn lại rồi đóng kết nối.

### 6. Cách chạy (`main.go` + `run.ps1`)

```powershell
# Cách 1: Chạy thủ công từng process
go run main.go <số_process> <id_process>

# Ví dụ: chạy 2 process
go run main.go 2 0    # Terminal 1
go run main.go 2 1    # Terminal 2

# Cách 2: Dùng script
.\run.ps1
# Nhập số process, script tự tạo N process
```

### 7. Logging & Kết quả

Mỗi process tạo 2 file log:
- **`XX__sender.log`**: Ghi lại mỗi lần gửi tin nhắn, bao gồm:
  - Thời gian gửi
  - Sender ID, Receiver ID
  - Nội dung tin nhắn
  - Trạng thái clock tại thời điểm gửi (logical clock + kiến thức về các process khác)

- **`XX__receiver.log`**: Ghi lại mỗi lần nhận tin nhắn, bao gồm:
  - Thời gian nhận
  - Sender ID, Receiver ID
  - Nội dung tin nhắn
  - `T_m` (timestamp người gửi gán cho người nhận)
  - Clock hiện tại của người nhận
  - **Status**: `delivering` (deliver ngay), `buffered` (đưa vào queue), hoặc `delivering from buffer` (deliver từ queue)
  - **Delivery Condition**: So sánh `tP_rcv > T_m`

---

## Ví dụ hoạt động (từ log thực tế)

Với 2 process (P0 và P1), ta thấy trong receiver log của P0:

1. **Message 7 từ P1**: Đến **sau** message 8, 9 → Message 8, 9 bị **buffered**.
2. **Message 7 deliver trước**, sau đó message 10 đến → message 10 cũng bị **buffered**.
3. Hệ thống kiểm tra queue → message 8 đủ điều kiện → **deliver from buffer**.
4. Tiếp tục → message 9 đủ điều kiện → **deliver from buffer**.
5. Tiếp tục → message 10 đủ điều kiện → **deliver from buffer**.

Kết quả: Dù tin nhắn đến theo thứ tự 9 → 8 → 7 → 10, SES đảm bảo deliver theo thứ tự nhân quả: 7 → 8 → 9 → 10.

---

## Tổng kết

| Thành phần | Mô tả |
|---|---|
| **Lĩnh vực** | Distributed Systems - Causal Message Ordering |
| **Thuật toán** | SES (Schiper-Eggli-Sandoz, 1989) |
| **Ngôn ngữ** | Go 1.20 |
| **Giao thức mạng** | TCP trên localhost |
| **Cấu trúc dữ liệu** | Vector Clock (ma trận N×N Logical Clock) |
| **Mô phỏng lỗi mạng** | Delay ngẫu nhiên (100-1000ms) + đảo thứ tự tin nhắn |
| **Cơ chế đảm bảo** | Buffer tin nhắn đến sớm, deliver khi đủ điều kiện nhân quả |
| **Logging** | Ghi chi tiết sender/receiver log cho từng process |
| **Đã test** | 2 process và 7 process |
